package main

import "github.com/alufers/notifier/notifier"

//go:generate swag init

func main() {
	notifier.Run()
}
