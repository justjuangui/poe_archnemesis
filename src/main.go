package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/spf13/viper"
	"trompetin17.com/poe/src/config"
	"trompetin17.com/poe/src/helpers"
	"trompetin17.com/poe/src/ui"

	"gocv.io/x/gocv"
)

const title string = "ArchNemesis Tool by trompetin17"

var (
	basePath   string
	dataPath   string
	outputPath string

	inventoryOutputPath string
	resolveOutputPath   string
	needsOutputPath     string

	myBag       config.ArchNemesisBag
	data        config.DataDescription
	recipesDict config.ArchNemesisRecipe

	currentRecipe string
	currentAction int      = 0
	actions       []string = []string{"My Inventary", "What can i do?"}
	totalActions  int      = len(actions) - 1
)

func init() {
	basePath, _ = getCurrentFolder()

	dataPath = filepath.Join(basePath, "data")
	outputPath = filepath.Join(basePath, "output")

	inventoryOutputPath = filepath.Join(outputPath, "inventory.txt")
	resolveOutputPath = filepath.Join(outputPath, "resolve.txt")
	needsOutputPath = filepath.Join(outputPath, "needs.txt")

	err := loadConfig(dataPath)

	if err != nil {
		fmt.Println(err)
		return
	}
}

func getCurrentFolder() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}

	exPath := filepath.Dir(ex)
	return exPath, nil
}

func loadConfig(configPath string) error {
	viper.SetConfigName("data")
	viper.AddConfigPath(configPath)
	viper.SetConfigType("json")

	if err := viper.ReadInConfig(); err != nil {
		err1 := fmt.Errorf("config: Error reading config file, %s", err)
		return err1
	}

	viper.SetConfigName("my")
	viper.MergeInConfig()

	err := viper.Unmarshal(&data)
	if err != nil {
		err1 := fmt.Errorf("config: Unable to decode into struct, %s", err)
		return err1
	}

	// load recipes & calculate info
	recipesDict = make(config.ArchNemesisRecipe)
	for _, recipe := range data.Recipes {
		recipesDict[recipe.Id] = recipe.Ingredients
	}

	return nil
}

func printInConsole(g *gocui.Gui, s string) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("logs")
		if err != nil {
			return err
		}

		fmt.Fprintln(v, s)
		return nil
	})
}

func clearConsole(g *gocui.Gui) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("logs")
		if err != nil {
			return err
		}

		v.Clear()
		return nil
	})
}

func updateInventory(g *gocui.Gui) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("inventory")
		if err != nil {
			return err
		}
		v.Clear()

		for _, i := range myBag.ToMapString() {
			fmt.Fprintln(v, i)
		}
		return nil
	})
}

func updateWhatCanIDo(g *gocui.Gui, items []string) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("whatcanido")
		if err != nil {
			return err
		}
		v.Clear()

		for _, i := range items {
			fmt.Fprintln(v, i)
		}
		return nil
	})
}

func updateNeeds(g *gocui.Gui, needs []string) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("needs")
		if err != nil {
			return err
		}
		v.Clear()

		for _, i := range needs {
			fmt.Fprintln(v, i)
		}
		return nil
	})
}

func updateResolve(g *gocui.Gui, resolve []string) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("resolve")
		if err != nil {
			return err
		}
		v.Clear()

		for _, i := range resolve {
			fmt.Fprintln(v, i)
		}
		return nil
	})
}

func updateRecipes(g *gocui.Gui, recipes []string) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("recipes")
		if err != nil {
			return err
		}
		v.Clear()

		fmt.Fprintln(v, strings.Join(recipes, ","))

		return nil
	})
}

func getInventoryInfo(g *gocui.Gui) {
	// by default use image called mystash.png
	myStashPath := filepath.Join(basePath, "mystash.png")

	imgStash := gocv.IMRead(myStashPath, gocv.IMReadGrayScale)

	if imgStash.Empty() {
		printInConsole(g, fmt.Sprintf("Cant read your stash inventory in %s\n", myStashPath))
		return
	}
	defer imgStash.Close()

	imgStashTrace := gocv.IMRead(myStashPath, gocv.IMReadColor)
	defer imgStashTrace.Close()

	// read the inventory
	myBag = make(config.ArchNemesisBag)
	for _, resource := range data.Resources {
		myBag[resource.Id] = 0 // by default

		// TODO: validate src
		// TODO: validate src path exists

		srcPath := filepath.Join(dataPath, resource.Src)

		img := gocv.NewMat()
		defer img.Close()

		imgFind := gocv.IMRead(srcPath, gocv.IMReadGrayScale)
		if imgFind.Empty() {
			printInConsole(g, fmt.Sprintf("Resource %s: Image no found %s\n", resource.Id, resource.Src))
			continue
		}
		defer imgFind.Close()

		imgFindW := imgFind.Cols()
		imgFindH := imgFind.Rows()

		gocv.MatchTemplate(imgStash, imgFind, &img, gocv.TmCcoeffNormed, gocv.NewMat())

		imgFindStash := imgStashTrace.Clone()
		defer imgFindStash.Close()

		for row := 0; row < img.Rows(); row++ {
			for col := 0; col < img.Cols(); col++ {

				valueF := img.GetFloatAt(row, col)

				if valueF >= 0.98 {
					gocv.Rectangle(&imgFindStash, image.Rect(col, row, col+imgFindW, row+imgFindH), color.RGBA{255, 0, 0, 0}, 2)

					// add in stashBag
					myBag[resource.Id]++
				}
			}
		}

		// save image in output
		if data.Trace {
			outputImg := filepath.Join(outputPath, fmt.Sprintf("%s.png", resource.Id))
			printInConsole(g, fmt.Sprintf("writing trace for %s: %s\n", resource.Id, outputImg))
			if ok := gocv.IMWrite(outputImg, imgFindStash); !ok {
				printInConsole(g, fmt.Sprintf("Failed for %s: %s\n", resource.Id, outputImg))
			}
		}
	}

	writeToFile(g, inventoryOutputPath, myBag.ToMapString())
	go updateInventory(g)
}

func calculateMyRecipes(g *gocui.Gui) {
	// evaluate and expand
	needs := make(config.ArchNemesisBag)
	messages := []string{}

	myBagClone := myBag.Clone()
	if currentRecipe == "" {
		currentRecipe = data.RecipeIWant
	}

	go updateRecipes(g, []string{currentRecipe})
	helpers.Calculate(&messages, &needs, &myBagClone, &recipesDict, currentRecipe, 1, 0, false)

	writeToFile(g, resolveOutputPath, messages)
	writeToFile(g, needsOutputPath, needs.ToMapString())

	go updateNeeds(g, needs.ToMapString())
	go updateResolve(g, messages)
}

func calculateWhatCanIBuild(g *gocui.Gui) {
	// Evaluate if I can build any Recipe
	builds := []string{}
	for k, v := range recipesDict {
		if len(v) == 0 {
			continue
		}

		tmpNeeds := make(config.ArchNemesisBag)
		tmpMessages := []string{}

		myBagClone := myBag.Clone()
		helpers.Calculate(&tmpMessages, &tmpNeeds, &myBagClone, &recipesDict, k, 1, 0, true)

		if len(tmpNeeds) == 0 {
			// A Candidate to build
			builds = append(builds, fmt.Sprintf("%-28s", k))
			writeToFile(g, filepath.Join(outputPath, fmt.Sprintf("CAN BUILD %s.txt", k)), tmpMessages)
		}
	}
	sort.Strings(builds)
	go updateWhatCanIDo(g, builds)
}

func processInfo(g *gocui.Gui) {
	clearConsole(g)
	err := loadConfig(basePath)

	if err != nil {
		printInConsole(g, fmt.Sprintln("Can't open config", err))
		return
	}

	if data.Trace {

		if _, err := os.Stat(outputPath); err == nil {
			if err := os.RemoveAll(outputPath); err != nil {
				printInConsole(g, fmt.Sprintf("Folder %s couldn't delete\n", outputPath))
				return
			}
		}
		err := os.Mkdir(outputPath, 0755)
		if err != nil {
			printInConsole(g, fmt.Sprintf("Folder %s couldn't create\n", outputPath))
			return
		}
	}

	getInventoryInfo(g)
	go calculateMyRecipes(g)
	go calculateWhatCanIBuild(g)
}

func openOuputFolder(g *gocui.Gui, v *gocui.View) error {
	cmd := exec.Command("explorer", outputPath)
	cmd.Run()

	return nil
}

func reProcessInfo(g *gocui.Gui, v *gocui.View) error {
	go processInfo(g)
	return nil
}

func screenshot(g *gocui.Gui, v *gocui.View) error {
	//take screenshot poe
	img, err := helpers.Capture()
	if err != nil {
		printInConsole(g, fmt.Sprintln("can't capture POE window", err))
		return nil
	}

	mtaImg, err := gocv.ImageToMatRGBA(img)
	if err != nil {
		printInConsole(g, fmt.Sprintln("can't load himage from POE", err))
		return nil
	}

	mtaFlip := gocv.NewMat()
	gocv.Flip(mtaImg, &mtaFlip, 0)

	//save to file both
	gocv.IMWrite("mystash.png", mtaFlip)

	go processInfo(g)
	return nil
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	if !ui.CanChangeView {
		return nil
	}

	nextIndex := (ui.CurrentIndexView + 1) % len(ui.Views)
	name := ui.Views[nextIndex]

	if _, err := ui.SetCurrentViewOnTop(g, name); err != nil {
		return err
	}

	ui.CurrentIndexView = nextIndex
	return nil
}

func updateActions(g *gocui.Gui) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("actions")
		if err != nil {
			return err
		}

		v.Clear()
		for i, a := range actions {
			selected := ""
			if i == currentAction {
				selected = "X"
			}
			fmt.Fprintf(v, "[%1s] %-20s ", selected, a)
		}
		return nil
	})
	if currentAction == 0 {
		g.SetViewOnTop("inventory")
	} else {
		g.SetViewOnTop("whatcanido")
	}
}

func actionsLeft(g *gocui.Gui, v *gocui.View) error {
	nextAction := currentAction - 1
	if nextAction < 0 {
		nextAction = 0
	}

	if nextAction != currentAction {
		currentAction = nextAction
		go updateActions(g)
	}
	return nil
}

func actionsRight(g *gocui.Gui, v *gocui.View) error {
	nextAction := (currentAction + 1)
	if nextAction > totalActions {
		nextAction = totalActions
	}

	if nextAction != currentAction {
		currentAction = nextAction
		go updateActions(g)
	}
	return nil
}

func recipeEnter(g *gocui.Gui, v *gocui.View) error {
	g.Update(func(g *gocui.Gui) error {
		v, err := ui.ShowRecipes(g)
		if err != nil {
			return err
		}

		// fill the information
		recipes := recipesDict.ToMapString()
		v.Title = fmt.Sprintf("Select Recipe you want (%d)", len(recipes))

		ox, oy := v.Origin()

		for i, r := range recipes {
			fmt.Fprintln(v, r)
			if r == currentRecipe {
				oy += i
			}
		}

		v.SetCursor(ox, oy)
		return nil
	})

	return nil
}

func showRecipeEsc(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		_, cy := v.Cursor()
		_, oy := v.Origin()
		curRecipeIndex := oy + cy
		newCurrentRecipe := recipesDict.ToMapString()[curRecipeIndex]
		if currentRecipe != newCurrentRecipe {
			currentRecipe = newCurrentRecipe
			go calculateMyRecipes(g)
		}
	}
	_, err := ui.HideRecipes(g)
	return err
}

func showRecipeCursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func showRecipeCursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func initKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlF, gocui.ModNone, screenshot); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlR, gocui.ModNone, reProcessInfo); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlO, gocui.ModNone, openOuputFolder); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		return err
	}

	// Actions handlers
	if err := g.SetKeybinding("actions", gocui.KeyArrowLeft, gocui.ModNone, actionsLeft); err != nil {
		return err
	}

	if err := g.SetKeybinding("actions", gocui.KeyArrowRight, gocui.ModNone, actionsRight); err != nil {
		return err
	}

	// recipe actions
	if err := g.SetKeybinding("recipes", gocui.KeyEnter, gocui.ModNone, recipeEnter); err != nil {
		return err
	}

	// show recipe actions
	if err := g.SetKeybinding("showrecipes", gocui.KeyEnter, gocui.ModNone, showRecipeEsc); err != nil {
		return err
	}
	if err := g.SetKeybinding("showrecipes", gocui.KeyArrowDown, gocui.ModNone, showRecipeCursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("showrecipes", gocui.KeyArrowUp, gocui.ModNone, showRecipeCursorUp); err != nil {
		return err
	}

	return nil
}

func main() {
	fmt.Println(title)

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		fmt.Println("Can't create interface", err)
		fmt.Scanln()
		return
	}
	defer g.Close()

	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen

	g.SetManagerFunc(ui.Layout)

	if err := initKeybindings(g); err != nil {
		fmt.Println("Can't bind quit", err)
		fmt.Scanln()
		return
	}

	updateActions(g)
	if currentRecipe == "" {
		currentRecipe = data.RecipeIWant
	}
	go updateRecipes(g, []string{currentRecipe})

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		fmt.Println("Error exiting app", err)
		fmt.Scanln()
		return
	}
}

func writeToFile(g *gocui.Gui, filePath string, d []string) {
	f, err := os.Create(filePath)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}

	for _, v := range d {
		fmt.Fprintln(f, v)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	printInConsole(g, fmt.Sprintf("file %s written successfully\n", filePath))
}
