package main

import "github.com/ChrisMcKee/hanu"

var commandList []hanu.CommandInterface
var dialogInteractions []hanu.DialogCfg

var Bot *hanu.Bot

// Register adds a new command to commandList
func Register(cmd string, description string, handler hanu.Handler) {
	commandList = append(commandList, hanu.NewCommand(cmd, description, handler))
}

func RegisterDialogInteraction(cfg hanu.DialogCfg) {
	dialogInteractions = append(dialogInteractions, cfg)
}

// List returns commandList
func List() []hanu.CommandInterface {
	return commandList
}

func ListDialogInteractions() []hanu.DialogCfg {
	return dialogInteractions
}
