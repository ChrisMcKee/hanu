package main

import "github.com/ChrisMcKee/hanu"

var commandList []hanu.CommandInterface

var Bot *hanu.Bot

// Register adds a new command to commandList
func Register(cmd string, description string, handler hanu.Handler) {
	commandList = append(commandList, hanu.NewCommand(cmd, description, handler))
}

// List returns commandList
func List() []hanu.CommandInterface {
	return commandList
}
