package main

import "os"

import "application"

func main() {
	app := application.App{}
	app.Initialize(
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("AUTH_USER"),
		os.Getenv("AUTH_PASSWORD"))
	app.Run(os.Getenv("PORT"))
}
