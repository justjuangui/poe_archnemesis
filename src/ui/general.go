package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

const (
	margin int = 1
)

func Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("actions", 0, 0, maxX-margin, margin*2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Actions"
		fmt.Fprintf(v, "[%1s] %-20s [%1s] %-20s", "X", "My Recipes", "", "What can i do?")
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
