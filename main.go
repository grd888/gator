package main

import (
	"os"
	"fmt"
	"github.com/grd888/gator/internal/config"
	"github.com/grd888/gator/internal/commands"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}
	// Initialize application state
	_ = commands.State{Config: &cfg}
	c := commands.NewCommands()
	c.Register("login", commands.HandlerLogin)
	// get the command from the command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: gator <command> [args]")
		os.Exit(1)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	// create the command instance
	command := commands.Command{Name: cmd, Args: args}
	// run the command
	err = c.Run(&commands.State{Config: &cfg}, command)
	if err != nil {
		fmt.Println("Error running command:", err)
		os.Exit(1)
	}
}