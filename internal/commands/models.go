package commands

import (
	"errors"
	"github.com/grd888/gator/internal/config"
)

// State holds the application state including configuration
type State struct {
	Config *config.Config
}

type Command struct {
	Name string
	Args []string

}

type Commands struct {
	commandsMap map[string]func(state *State, cmd Command) error 
}

// NewCommands creates a new Commands instance with initialized map
func NewCommands() *Commands {
	return &Commands{
		commandsMap: make(map[string]func(state *State, cmd Command) error),
	}
}

func (c *Commands) Run(state *State, cmd Command) error {
	if handler, ok := c.commandsMap[cmd.Name]; ok {
		return handler(state, cmd)
	}
	return errors.New("command not found")
}

func (c *Commands) Register(name string, handler func(state *State, cmd Command) error) {
	c.commandsMap[name] = handler
}
