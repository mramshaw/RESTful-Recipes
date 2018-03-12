package main

import "os"

import "application"

func main() {
	app := application.App{}
	app.Initialize(
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))
	app.Run(os.Getenv("PORT"))
}
