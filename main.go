package main

import (
	"database/sql"
	"fmt"
	"os"
	
	"github.com/grd888/gator/internal/commands"
	"github.com/grd888/gator/internal/config"
	"github.com/grd888/gator/internal/database"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	// First read the configuration to get the database URL
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}
	
	// Connect to the database using the URL from config
	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()
	
	// Create a queries object for database operations
	dbQueries := database.New(db)
	
	// Initialize application state with config and database access
	appState := commands.State{
		Config: &cfg,
		DB:     dbQueries,
	}
	c := commands.NewCommands()
	c.Register("login", commands.HandlerLogin)
	c.Register("register", commands.HandlerRegister)
	c.Register("reset", commands.HandlerReset)
	c.Register("users", commands.HandlerListUsers)
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
	err = c.Run(&appState, command)
	if err != nil {
		fmt.Println("Error running command:", err)
		os.Exit(1)
	}
}