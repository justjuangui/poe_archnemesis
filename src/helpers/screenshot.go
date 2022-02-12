package helpers

import (
	"fmt"
	"image"
	"reflect"
	"syscall"
	"unsafe"
)

var (
	modUser32         = syscall.NewLazyDLL("User32.dll")
	procFindWindow    = modUser32.NewProc("FindWindowW")
	procGetClientRect = modUser32.NewProc("GetClientRect")
	procGetDC         = modUser32.NewProc("GetDC")
	procReleaseDC     = modUser32.NewProc("ReleaseDC")

	modGdi32                   = syscall.NewLazyDLL("Gdi32.dll")
	procBitBlt                 = modGdi32.NewProc("BitBlt")
	procCreateCompatibleBitmap = modGdi32.NewProc("CreateCompatibleBitmap")
	procCreateCompatibleDC     = modGdi32.NewProc("CreateCompatibleDC")
	procCreateDIBSection       = modGdi32.NewProc("CreateDIBSection")
	procDeleteDC               = modGdi32.NewProc("DeleteDC")
	procDeleteObject           = modGdi32.NewProc("DeleteObject")
	procGetDeviceCaps          = modGdi32.NewProc("GetDeviceCaps")
	procSelectObject           = modGdi32.NewProc("SelectObject")

	modShcore                  = syscall.NewLazyDLL("Shcore.dll")
	procSetProcessDpiAwareness = modShcore.NewProc("SetProcessDpiAwareness")
)

const (
	// GetDeviceCaps constants from Wingdi.h
	deviceCaps_HORZRES    = 8
	deviceCaps_VERTRES    = 10
	deviceCaps_LOGPIXELSX = 88
	deviceCaps_LOGPIXELSY = 90

	// BitBlt constants
	bitBlt_SRCCOPY = 0x00CC0020
)

// Windows RECT structure
type win_RECT struct {
	Left, Top, Right, Bottom int32
}

// http://msdn.microsoft.com/en-us/library/windows/desktop/dd183375.aspx
type win_BITMAPINFO struct {
	BmiHeader win_BITMAPINFOHEADER
	BmiColors *win_RGBQUAD
}

// http://msdn.microsoft.com/en-us/library/windows/desktop/dd183376.aspx
type win_BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

// http://msdn.microsoft.com/en-us/library/windows/desktop/dd162938.aspx
type win_RGBQUAD struct {
	RgbBlue     byte
	RgbGreen    byte
	RgbRed      byte
	RgbReserved byte
}

func init() {
	procSetProcessDpiAwareness.Call(uintptr(2)) // PROCESS_PER_MONITOR_DPI_AWARE
}

func findWindow() (syscall.Handle, error) {
	var handle syscall.Handle

	stringPoint, err := syscall.UTF16PtrFromString("Path of Exile")
	if err != nil {
		return handle, err
	}

	ret, _, _ := procFindWindow.Call(
		0, uintptr(unsafe.Pointer(stringPoint)),
	)

	if ret == 0 {
		return handle, fmt.Errorf("not found POE , Is it running POE")
	}

	handle = syscall.Handle(ret)
	return handle, nil
}

func windowRect(hwnd syscall.Handle) (image.Rectangle, error) {
	var rect win_RECT
	ret, _, err := procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rect)))
	if ret == 0 {
		return image.Rectangle{}, fmt.Errorf("error getting window dimensions: %s", err)
	}

	return image.Rect(0, 0, int(rect.Right), int(rect.Bottom)), nil
}

func Capture() (image.Image, error) {
	handle, err := findWindow()
	if err != nil {
		return nil, err
	}

	rect, err := windowRect(handle)
	if err != nil {
		return nil, err
	}

	return captureWindow(handle, rect)
}

func captureWindow(handle syscall.Handle, rect image.Rectangle) (image.Image, error) {
	dcSrc, _, err := procGetDC.Call(uintptr(handle))
	if dcSrc == 0 {
		return nil, fmt.Errorf("error preparing screen capture: %s", err)
	}
	defer procReleaseDC.Call(0, dcSrc)

	dcDst, _, err := procCreateCompatibleDC.Call(dcSrc)
	if dcDst == 0 {
		return nil, fmt.Errorf("error creating DC for drawing: %s", err)
	}
	defer procDeleteDC.Call(dcDst)

	width := rect.Dx() / 2 // only take half of the window
	height := rect.Dy()

	var bitmapInfo win_BITMAPINFO

	bitmapInfo.BmiHeader = win_BITMAPINFOHEADER{
		BiSize:        uint32(reflect.TypeOf(bitmapInfo.BmiHeader).Size()),
		BiWidth:       int32(width),
		BiHeight:      int32(height),
		BiPlanes:      1,
		BiBitCount:    32,
		BiCompression: 0, // BI_RGB
	}
	bitmapData, _, err := procCreateCompatibleBitmap.Call(dcDst, uintptr(width), uintptr(height))
	if dcDst == 0 {
		return nil, fmt.Errorf("error creating bitmap for drawing: %s", err)
	}

	bitmap, _, err := procCreateDIBSection.Call(
		dcDst,
		uintptr(unsafe.Pointer(&bitmapInfo)),
		0,
		uintptr(unsafe.Pointer(&bitmapData)),
		0,
		0,
	)

	if bitmap == 0 {
		return nil, fmt.Errorf("error creating bitmap for screen capture: %s", err)
	}
	defer procDeleteObject.Call(bitmap)

	hOld, _, _ := procSelectObject.Call(dcDst, bitmap)
	ret, _, err := procBitBlt.Call(
		dcDst, 0, 0, uintptr(width), uintptr(height), dcSrc, uintptr(rect.Min.X), uintptr(rect.Min.Y), bitBlt_SRCCOPY,
	)
	if ret == 0 {
		return nil, fmt.Errorf("error capturing screen: %s", err)
	}
	procSelectObject.Call(dcDst, hOld)

	var slice []byte
	sliceHdr := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	sliceHdr.Data = uintptr(bitmapData)
	sliceHdr.Len = width * height * 4
	sliceHdr.Cap = sliceHdr.Len

	imageBytes := make([]byte, len(slice))
	for i := 0; i < len(imageBytes); i += 4 {
		imageBytes[i], imageBytes[i+2], imageBytes[i+1], imageBytes[i+3] = slice[i+2], slice[i], slice[i+1], slice[i+3]
	}

	img := &image.RGBA{imageBytes, 4 * width, image.Rect(0, 0, width, height)}
	return img, nil
}
