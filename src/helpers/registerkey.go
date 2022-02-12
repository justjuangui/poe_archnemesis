package helpers

import (
	"bytes"
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

const (
	ModAlt = 1 << iota
	ModCtrl
	ModShift
	ModWin
)

type Hotkey struct {
	Id          int
	Modifiers   int
	KeyCode     int
	Description string
}

type MSG struct {
	HWND   uintptr
	UINT   uintptr
	WPARAM int16
	LPARAM int64
	DWORD  int32
	POINT  struct{ X, Y int64 }
}

func (h *Hotkey) String() string {
	mod := &bytes.Buffer{}
	if h.Modifiers&ModAlt != 0 {
		mod.WriteString("Alt+")
	}

	if h.Modifiers&ModCtrl != 0 {
		mod.WriteString("Ctrl+")
	}

	if h.Modifiers&ModShift != 0 {
		mod.WriteString("Shift+")
	}

	if h.Modifiers&ModWin != 0 {
		mod.WriteString("Win+")
	}

	return fmt.Sprintf("Hotkey[Id: %d, %s%c] %s", h.Id, mod, h.KeyCode, h.Description)
}

func RegisterKey(handlerChannel chan<- int) {
	user32 := syscall.MustLoadDLL("user32")
	defer user32.Release()

	reghotkey := user32.MustFindProc("RegisterHotKey")

	keys := map[int16]*Hotkey{
		1: {1, ModAlt + ModCtrl, 'F', "Capture POE Window"},
		2: {2, ModAlt + ModCtrl, 'X', "Close Application"},
	}

	for _, v := range keys {
		r1, _, err := reghotkey.Call(
			0, uintptr(v.Id), uintptr(v.Modifiers), uintptr(v.KeyCode),
		)
		if r1 == 1 {
			fmt.Println("Registered", v)
		} else {
			fmt.Println("Failed to register", v, ", error;", err)
		}
	}

	peekmsg := user32.MustFindProc("PeekMessageW")
	for {
		var msg = &MSG{}
		peekmsg.Call(uintptr(unsafe.Pointer(msg)), 0, 0, 0, 1)

		if id := msg.WPARAM; id != 0 {
			fmt.Println("Hotkey pressed:", keys[id])
			if id == 2 {
				fmt.Println("Goodbye")
				break
			} else if id == 1 {
				handlerChannel <- 1 // Capture screen
				fmt.Println("Continue")
			}
		}
		time.Sleep(time.Millisecond * 50)
	}

	handlerChannel <- 0 //QUIT
}
