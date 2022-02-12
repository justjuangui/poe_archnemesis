package config

type ResourceDescription struct {
	Id  string `json,mapstructure:"id"`
	Src string `json,mapstructure:"src"`
}

type RecipeDescription struct {
	Id          string   `json,mapstructure:"id"`
	Ingredients []string `json,mapstructure:"ingredients"`
}

type DataDescription struct {
	Trace       bool                  `json,mapstructure:"trace"`
	RecipeIWant string                `json,mapstructure:"recipesIWant"`
	Resources   []ResourceDescription `json,mapstructure:"resources"`
	Recipes     []RecipeDescription   `json,mapstructure:"recipes"`
}
