package actions

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"

	"simplecms/models"
)

// RegisterGet displays the registration form
func RegisterGet(c buffalo.Context) error {
	// Make empty user available to the registration form
	c.Set("user", &models.User{})
	return c.Render(http.StatusOK, r.HTML("auth/register.plush.html"))
}

// RegisterPost processes the registration form
func RegisterPost(c buffalo.Context) error {
	// Allocate an empty user
	user := &models.User{}

	// Bind the user to the request body
	if err := c.Bind(user); err != nil {
		return errors.New("could not parse registration form")
	}

	// Get the DB connection from context
	tx := c.Value("tx").(*pop.Connection)

	// Create a new validator
	verrs := validate.NewErrors()

	// Check if password is present
	if user.Password == "" {
		verrs.Add("password", "Password is required")
	} else if len(user.Password) < 8 {
		verrs.Add("password", "Password must be at least 8 characters")
	}

	// Check password confirmation
	pwdConfirm := c.Request().FormValue("password_confirmation")
	if user.Password != pwdConfirm {
		verrs.Add("password_confirmation", "Passwords do not match")
	}

	// Hash the password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("could not encrypt password")
	}
	user.PasswordHash = string(passwordHash)

	// Validate and create user
	userVerrs, err := tx.ValidateAndCreate(user)
	if err != nil {
		// Check for unique email constraint
		if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "email") {
			verrs.Add("email", "Email address is already in use")
		} else {
			return errors.New("error creating user: " + err.Error())
		}
	}

	// Combine validation errors
	verrs.Append(userVerrs)

	if verrs.HasAny() {
		// If there are validation errors, render the registration form again
		c.Set("user", user)
		c.Set("errors", verrs)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("auth/register.plush.html"))
	}

	// Store the logged-in user ID in the session
	c.Session().Set("current_user_id", user.ID)
	c.Session().Save()

	// Set a flash message
	c.Flash().Add("success", "Registration successful. Welcome!")

	// Redirect to the home page
	return c.Redirect(http.StatusSeeOther, "/")
}

// LoginGet displays the login form
func LoginGet(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("auth/login.plush.html"))
}

// LoginPost processes the login form
func LoginPost(c buffalo.Context) error {
	// Get login credentials from the form
	email := c.Request().FormValue("email")
	password := c.Request().FormValue("password")

	// Validate form input
	verrs := validate.NewErrors()
	if email == "" {
		verrs.Add("email", "Email is required")
	}
	if password == "" {
		verrs.Add("password", "Password is required")
	}

	if verrs.HasAny() {
		c.Set("errors", verrs)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("auth/login.plush.html"))
	}

	// Get the DB connection from context
	tx := c.Value("tx").(*pop.Connection)

	// Find the user by email
	user := &models.User{}
	err := tx.Where("email = ?", email).First(user)

	// If user not found or password incorrect
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		verrs.Add("login", "Invalid email or password")
		c.Set("errors", verrs)
		return c.Render(http.StatusUnauthorized, r.HTML("auth/login.plush.html"))
	}

	// Store the logged-in user ID in the session
	c.Session().Set("current_user_id", user.ID)
	c.Session().Save()

	// Set a flash message
	c.Flash().Add("success", fmt.Sprintf("Welcome back, %s!", user.Name))

	// Redirect to home page
	return c.Redirect(http.StatusSeeOther, "/")
}

// Logout clears the session and logs out the user
func Logout(c buffalo.Context) error {
	// Clear the session
	c.Session().Clear()
	c.Session().Save()

	// Set a flash message
	c.Flash().Add("success", "You have been logged out")

	// Redirect to the home page
	return c.Redirect(http.StatusSeeOther, "/")
}

// SetCurrentUser attempts to find a user based on the current_user_id
// in the session. If one is found it is set on the context.
func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		// Skip if no session is defined
		if c.Session() == nil {
			return next(c)
		}

		// Get the user ID from the session
		userID := c.Session().Get("current_user_id")
		if userID == nil {
			return next(c)
		}

		// Try to parse the user ID as a UUID
		uid, ok := userID.(uuid.UUID)
		if !ok {
			// Try to convert string to UUID
			uidStr, ok := userID.(string)
			if !ok {
				return next(c)
			}
			
			var err error
			uid, err = uuid.FromString(uidStr)
			if err != nil {
				return next(c)
			}
		}

		// Get the DB connection from the context
		tx := c.Value("tx").(*pop.Connection)

		// Find the user by ID
		user := &models.User{}
		err := tx.Find(user, uid)
		if err != nil {
			c.Session().Delete("current_user_id")
			c.Session().Save()
			return next(c)
		}

		// Set the user on the context
		c.Set("current_user", user)
		return next(c)
	}
}

// Authorize requires a user to be logged in before accessing a route
func Authorize(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		// Get the current user from the context
		user := c.Value("current_user")
		if user == nil {
			// If not logged in, set a flash message
			c.Flash().Add("danger", "You must be logged in to view that page")
			
			// Return to the login page
			return c.Redirect(http.StatusFound, "/login")
		}

		// User is logged in, call the next handler
		return next(c)
	}
} 