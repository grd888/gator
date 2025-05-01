package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	
	"github.com/google/uuid"
	"github.com/grd888/gator/internal/database"
)

// parsePublishedDate attempts to parse a date string in various formats
func parsePublishedDate(dateStr string) (time.Time, error) {
	// List of time formats to try
	formats := []string{
		time.RFC1123Z,     // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,      // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC822Z,      // "02 Jan 06 15:04 -0700"
		time.RFC822,       // "02 Jan 06 15:04 MST"
		time.RFC3339,      // "2006-01-02T15:04:05Z07:00"
		time.RFC3339Nano,  // "2006-01-02T15:04:05.999999999Z07:00"
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02 Jan 2006",
		"January 2, 2006",
		"Jan 2, 2006",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05 MST",
	}

	// Try each format
	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse date: %s", dateStr)
}
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

func scrapeFeeds(state *State) error {
	// Get the next feed to fetch from the DB
	feed, err := state.DB.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("error getting next feed to fetch: %w", err)
	}

	// Mark it as fetched
	now := time.Now()
	_, err = state.DB.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{Time: now, Valid: true},
		ID:            feed.ID,
	})
	if err != nil {
		return fmt.Errorf("error marking feed as fetched: %w", err)
	}

	// Fetch the feed using the URL
	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("error fetching RSS feed: %w", err)
	}

	// Iterate over the items in the feed and save them to the database
	fmt.Printf("Feed: %s (%d items)\n", feed.Name, len(rssFeed.Channel.Item))
	for _, item := range rssFeed.Channel.Item {
		// Parse the published date
		var publishedAt sql.NullTime
		if item.PubDate != "" {
			// Try different time formats
			parsedTime, err := parsePublishedDate(item.PubDate)
			if err != nil {
				fmt.Printf("Warning: Could not parse published date '%s': %v\n", item.PubDate, err)
			} else {
				publishedAt = sql.NullTime{Time: parsedTime, Valid: true}
			}
		}

		// Create a new post
		postID := uuid.New()
		_, err := state.DB.CreatePost(context.Background(), database.CreatePostParams{
			ID:          postID,
			CreatedAt:   now,
			UpdatedAt:   now,
			FeedID:      feed.ID,
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: publishedAt,
		})

		if err != nil {
			// If the post already exists, just ignore the error
			if strings.Contains(err.Error(), "unique constraint") {
				// Post already exists, skip it
				continue
			}
			// Log other errors but continue processing
			fmt.Printf("Error saving post '%s': %v\n", item.Title, err)
		} else {
			fmt.Printf("* Saved: %s\n", item.Title)
		}
	}
	fmt.Println()

	return nil
}

func HandlerAgg(state *State, cmd Command, user database.User) error {
	// Check if we have the correct number of arguments
	if len(cmd.Args) < 1 {
		return errors.New("time_between_reqs is required (e.g. 1s, 1m, 1h)")
	}

	// Parse the time_between_reqs argument
	timeBetweenReqs := cmd.Args[0]
	timeBetweenRequests, err := time.ParseDuration(timeBetweenReqs)
	if err != nil {
		return fmt.Errorf("invalid time duration: %w", err)
	}

	// Print a message indicating the collection interval
	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests)

	// Create a ticker to run the scrapeFeeds function at regular intervals
	ticker := time.NewTicker(timeBetweenRequests)

	// Run the scrapeFeeds function immediately and then every time the ticker ticks
	for ; ; <-ticker.C {
		err := scrapeFeeds(state)
		if err != nil {
			fmt.Printf("Error scraping feeds: %v\n", err)
		}
	}
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

func HandlerBrowse(state *State, cmd Command, user database.User) error {
	// Default limit is 2 if not provided
	limit := int32(2)
	
	// Check if we have a limit argument
	if len(cmd.Args) > 0 {
		// Try to parse the limit
		n, err := fmt.Sscanf(cmd.Args[0], "%d", &limit)
		if err != nil || n != 1 {
			return fmt.Errorf("invalid limit: %s", cmd.Args[0])
		}
	}
	
	// Get posts for the current user
	posts, err := state.DB.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	})
	
	if err != nil {
		return fmt.Errorf("error getting posts: %w", err)
	}
	
	if len(posts) == 0 {
		fmt.Println("No posts found. Try following some feeds first.")
		return nil
	}
	
	fmt.Printf("Found %d posts:\n", len(posts))
	for _, post := range posts {
		// Format the published date if available
		publishedAt := "unknown date"
		if post.PublishedAt.Valid {
			publishedAt = post.PublishedAt.Time.Format("Jan 02, 2006")
		}
		
		fmt.Printf("\n* %s\n", post.Title)
		fmt.Printf("  URL: %s\n", post.Url)
		fmt.Printf("  Published: %s\n", publishedAt)
		
		// Print description if available
		if post.Description.Valid && post.Description.String != "" {
			// Truncate description if it's too long
			desc := post.Description.String
			if len(desc) > 100 {
				desc = desc[:100] + "..."
			}
			fmt.Printf("  Description: %s\n", desc)
		}
	}
	
	return nil
}
