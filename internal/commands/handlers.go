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

func HandlerListUsers(state *State, cmd Command, user database.User) error {
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

func HandlerAgg(state *State, cmd Command, user database.User) error {
	url := "https://www.wagslane.dev/index.xml"

	rssFeed, err := fetchFeed(context.Background(), url)
	if err != nil {
		return fmt.Errorf("error fetching RSS feed: %w", err)
	}
	fmt.Println(rssFeed)
	return nil
}

func HandlerAddFeed(state *State, cmd Command, user database.User) error {
	// Check if we have enough arguments
	if len(cmd.Args) < 2 {
		return errors.New("feed name and URL are required")
	}
	
	// Get the current user's ID from the database
	user, err := state.DB.GetUserByName(context.Background(), user.Name)
	if err != nil {
		return fmt.Errorf("error finding current user: %w", err)
	}
	
	name := cmd.Args[0]
	url := cmd.Args[1]
	
	// Validate the feed by fetching it
	rssFeed, err := fetchFeed(context.Background(), url)
	if err != nil {
		return fmt.Errorf("error fetching RSS feed: %w", err)
	}
	
	// Use the feed title from the RSS if available, otherwise use the provided name
	feedName := name
	if rssFeed.Channel.Title != "" {
		// If user provided a custom name, use it; otherwise use the feed's actual title
		if name == url || name == "" {
			feedName = rssFeed.Channel.Title
		}
	}
	
	// Create a new feed with the provided name and URL
	now := time.Now()
	feedID := uuid.New()
	
	createdFeed, err := state.DB.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        feedID,
		CreatedAt: now,
		UpdatedAt: now,
		Name:      feedName, // Use the feed title from RSS or user-provided name
		Url:       url,
		UserID:    user.ID, // Use the actual user ID from the database
	})
	
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			fmt.Println("Error: A feed with that name already exists")
			return err
		}
		fmt.Println("Error creating feed:", err)
		return err
	}
	
	// Automatically create a feed follow record for the current user
	feedFollowID := uuid.New()
	feedFollow, err := state.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        feedFollowID,
		CreatedAt: now,
		UpdatedAt: now,
		FeedID:    createdFeed.ID,
		UserID:    user.ID,
	})
	
	if err != nil {
		// If there's an error creating the feed follow, log it but don't fail the whole operation
		// since the feed was already created successfully
		fmt.Println("Warning: Feed created but could not be followed automatically:", err)
	} else {
		fmt.Printf("Feed created and followed successfully: %s\n", feedFollow.FeedName)
	}
	
	fmt.Printf("Title: %s\n", rssFeed.Channel.Title)
	fmt.Printf("Description: %s\n", rssFeed.Channel.Description)
	fmt.Printf("Items: %d\n", len(rssFeed.Channel.Item))		
	return nil	
}

func HandlerListFeeds(state *State, cmd Command, user database.User) error {
	feeds, err := state.DB.ListFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error listing feeds: %w", err)
	}
	for _, feed := range feeds {
		// get the user name from the user id
		user, err := state.DB.GetUser(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("error getting user: %w", err)
		}
		fmt.Printf("* %s (URL: %s) (User: %s)\n", feed.Name, feed.Url, user.Name)
	}
	return nil
}

func HandlerListFollowing(state *State, cmd Command, user database.User) error {
	// Get all feed follows for the current user
	feedFollows, err := state.DB.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error getting followed feeds: %w", err)
	}
	
	if len(feedFollows) == 0 {
		fmt.Println("You are not following any feeds.")
		return nil
	}
	
	fmt.Printf("Feeds followed by %s:\n", user.Name)
	for _, follow := range feedFollows {
		fmt.Printf("* %s\n", follow.FeedName)
	}
	
	return nil
}

func HandlerUnfollowFeed(state *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return errors.New("feed URL is required")
	}
	url := cmd.Args[0]

	// Delete the feed follow record
	err := state.DB.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		Url:    url,
	})
	if err != nil {
		return fmt.Errorf("error unfollowing feed: %w", err)
	}

	fmt.Printf("Successfully unfollowed feed: %s\n", url)
	return nil
}

func HandlerFollowFeed(state *State, cmd Command, user database.User) error {
	// Check if we have enough arguments
	if len(cmd.Args) < 1 {
		return errors.New("feed URL is required")
	}
	
	// Get the URL from arguments
	url := cmd.Args[0]
	
	// Check if the feed exists in our database
	feed, err := state.DB.GetFeedByUrl(context.Background(), url)
	if err != nil {
		// If the feed doesn't exist, return an error
		return fmt.Errorf("feed with URL %s not found: %w", url, err)
	}
	
	// Now create the feed follow
	now := time.Now()
	feedFollowID := uuid.New()
	
	feedFollow, err := state.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        feedFollowID,
		CreatedAt: now,
		UpdatedAt: now,
		FeedID:    feed.ID,
		UserID:    user.ID,
	})
	
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("you are already following this feed")
		}
		return fmt.Errorf("error following feed: %w", err)
	}
	
	fmt.Printf("Successfully followed feed: %s\n", feedFollow.FeedName)
	fmt.Printf("User: %s\n", feedFollow.UserName)
	return nil
}
