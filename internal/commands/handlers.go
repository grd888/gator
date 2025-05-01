package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	
	"github.com/google/uuid"
	"github.com/grd888/gator/internal/database"
)
// HandlerLogin handles the login command
func HandlerLogin(state *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("username is required")
	}
	// Check if the username exists in the database
	username := cmd.Args[0]
	_, err := state.DB.GetUserByName(context.Background(), username)
	if err != nil {
		return fmt.Errorf("user '%s' does not exist: %w", username, err)
	}

	state.Config.SetUser(cmd.Args[0])
	fmt.Println("User set to:", cmd.Args[0])
	return nil
}

func HandlerReset(state *State, cmd Command) error {
	err := state.DB.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error deleting users: %w", err)
	}
	fmt.Println("All users deleted")
	return nil
}

func HandlerRegister(state *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return errors.New("username is required")
	}
	username := cmd.Args[0]
	
	// Create a new user with the provided username
	now := time.Now()
	userID := uuid.New()
	
	_, err := state.DB.CreateUser(context.Background(), database.CreateUserParams{
		ID:        userID,
		CreatedAt: now,
		UpdatedAt: now,
		Name:      username,
	})
	
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			fmt.Println("Error: A user with that name already exists")
			return err
		}
		fmt.Println("Error creating user:", err)
		return err
	}
	
	// Set the current user in config
	err = state.Config.SetUser(username)
	if err != nil {
		fmt.Println("Error setting user in config:", err)
		return err
	}
	
	fmt.Printf("User created successfully: %s (ID: %s)\n", username, userID)


	return nil
}

func HandlerListUsers(state *State, cmd Command) error {
	users, err := state.DB.ListUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error listing users: %w", err)
	}
	for _, user := range users {
		// if user is the current user append (current)
		if user.Name == state.Config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func HandlerAgg(state *State, cmd Command) error {
	url := "https://www.wagslane.dev/index.xml"

	rssFeed, err := fetchFeed(context.Background(), url)
	if err != nil {
		return fmt.Errorf("error fetching RSS feed: %w", err)
	}
	fmt.Println(rssFeed)
	return nil
}

	
	