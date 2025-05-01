package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/grd888/gator/internal/database"
)

// MiddlewareLoggedIn is a higher-order function that wraps handlers requiring a logged-in user.
// It checks if a user is logged in, retrieves the user from the database, and passes it to the handler.
// This centralizes the authentication logic and reduces code duplication.
func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		// Check if a user is logged in
		currentUser := s.Config.CurrentUserName
		if currentUser == "" {
			return errors.New("no user set - please login first")
		}

		// Get the current user from the database
		user, err := s.DB.GetUserByName(context.Background(), currentUser)
		if err != nil {
			return fmt.Errorf("error finding current user: %w", err)
		}

		// Call the original handler with the user
		return handler(s, cmd, user)
	}
}