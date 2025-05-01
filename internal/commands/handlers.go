package commands

import (
	"errors"
	"fmt"
)
// HandlerLogin handles the login command
func HandlerLogin(state *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("username is required")
	}
	state.Config.SetUser(cmd.Args[0])
	fmt.Println("User set to:", cmd.Args[0])
	return nil
}
	
	