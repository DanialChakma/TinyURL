package main

import (
	"go.mod/initializers"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectDB()
}

func main() {
	// initializers.DB.AutoMigrate(&models.Post{})
}
