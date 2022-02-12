package helpers

import (
	"fmt"
	"strings"

	"trompetin17.com/poe/src/config"
)

func Calculate(messages *[]string, needs *config.ArchNemesisBag, myBag *config.ArchNemesisBag, recipes *config.ArchNemesisRecipe, recipe string, need int, deep int, evaluateOnlyNeed bool) {
	// find the recipe that user want
	_, ok := (*recipes)[recipe]
	if !ok {
		fmt.Printf("Recipe %s not found in data file\n", recipe)
		return
	}

	currentRecipe := (*recipes)[recipe]
	currentInv := (*myBag)[recipe]

	if deep == 0 || currentInv < need {
		if len(currentRecipe) == 0 {
			(*messages) = append((*messages), fmt.Sprintf("%-30s you need %2d", fmt.Sprintf("%s>%s", strings.Repeat("=", deep), recipe), need-currentInv))
			(*needs)[recipe] = (*needs)[recipe] + (need - currentInv)
			(*myBag)[recipe] = 0
		} else {
			// evaluate inventory for ingredients
			currentIngredientesInv := make(config.ArchNemesisBag)
			maxQuantity := -1
			for _, ing := range currentRecipe {
				quantity := (*myBag)[ing]
				currentIngredientesInv[ing] = quantity
				if quantity > maxQuantity {
					maxQuantity = quantity
				}
			}

			messageFormat := "%-30s you can build %2d"
			if maxQuantity == 0 {
				maxQuantity = need
				messageFormat = "%-30s to build %2d you need"
			} else if (maxQuantity > need && deep > 0) || evaluateOnlyNeed {
				maxQuantity = need
			}

			(*messages) = append((*messages), fmt.Sprintf(messageFormat, fmt.Sprintf("%s>%s", strings.Repeat("=", deep), recipe), maxQuantity))
			for _, ing := range currentRecipe {
				Calculate(messages, needs, myBag, recipes, ing, maxQuantity, deep+1, evaluateOnlyNeed)
			}

		}
	} else {
		(*myBag)[recipe] -= need
		(*messages) = append((*messages), fmt.Sprintf("%-30s you have %2d", fmt.Sprintf("%s>%s", strings.Repeat("=", deep), recipe), need))
	}
}
