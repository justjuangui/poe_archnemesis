package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

const (
	margin int = 1
)

var (
	CurrentIndexView int      = -1
	Views            []string = []string{"actions", "recipes"}
	CanChangeView    bool     = true
)

func SetCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("showrecipes", margin+31+margin, margin*6, maxX-margin-31-margin-margin, maxY-6-margin-margin); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Clear()
	}

	if v, err := g.SetView("actions", 0, 0, maxX-margin, margin*2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "View"
	}

	if v, err := g.SetView("whatcanido", 0, margin*3, margin+31, maxY-6-margin-margin); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "I can do"
	}

	if v, err := g.SetView("inventory", 0, margin*3, margin+31, maxY-6-margin-margin); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Inventory"
	}

	if v, err := g.SetView("recipes", margin+31+margin, margin*3, maxX-margin, margin*5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Recipes"
	}

	if v, err := g.SetView("resolve", margin+31+margin, margin*6, maxX-margin-31-margin-margin, maxY-6-margin-margin); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Resolve"
	}

	if v, err := g.SetView("needs", maxX-margin-31-margin, margin*6, maxX-margin, maxY-6-margin-margin); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Needs"
	}

	if v, err := g.SetView("logs", margin+31+margin, maxY-6-margin, maxX-margin, maxY-margin); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Logs"
		v.Autoscroll = true
	}

	if v, err := g.SetView("Keybindings", 0, maxY-6-margin, margin+31, maxY-margin); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		fmt.Fprintln(v, "^F: POE Print screen")
		fmt.Fprintln(v, "^O: Open ouput folder")
		fmt.Fprintln(v, "^R: Process mystash.png")
		fmt.Fprintln(v, "^C: Exit")
		v.Title = "Help"
	}
	return nil
}

func ShowRecipes(g *gocui.Gui) (*gocui.View, error) {
	g.Cursor = true
	CanChangeView = false
	return SetCurrentViewOnTop(g, "showrecipes")
}

func HideRecipes(g *gocui.Gui) (*gocui.View, error) {
	g.Cursor = false
	CanChangeView = true
	CurrentIndexView = 1 // recipes
	g.SetViewOnBottom("showrecipes")
	return SetCurrentViewOnTop(g, "recipes")
}
