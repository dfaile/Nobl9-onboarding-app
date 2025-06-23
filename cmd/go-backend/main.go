package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	"github.com/nobl9/nobl9-go/sdk"
)

// Valid roles that can be assigned
var validRoles = map[string]bool{
	"project-owner":  true,
	"project-viewer": true,
	"project-editor": true,
}

// UserGroup represents a group of users and their role
type UserGroup struct {
	UserIDs string `json:"userIds"` // Comma-separated list of user IDs or emails
	Role    string `json:"role"`    // Role to assign (must be one of validRoles)
}

// CreateProjectRequest defines the request payload for creating a project
type CreateProjectRequest struct {
	AppID       string      `json:"appID"`       // Name of the project to create
	Description string      `json:"description"` // Description of the project (optional)
	UserGroups  []UserGroup `json:"userGroups"`  // List of user groups with their roles
}

// Response defines the API response structure sent back to the client
type Response struct {
	Success bool   `json:"success"` // Whether the operation was successful
	Message string `json:"message"` // Human-readable message about the operation
}

// ptr creates a pointer to a string - helper function needed for role binding specs
// This is the same helper function from your working CLI tool
func ptr(s string) *string {
	return &s
}

// sanitizeName ensures the string is RFC-1123 compliant by converting to lowercase,
// replacing non-alphanumeric (except hyphen) characters with hyphens,
// and trimming leading/trailing hyphens.
// This is copied from your working CLI tool
func sanitizeName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	// Replace non-alphanumeric characters (except hyphen) with a hyphen
	reg := regexp.MustCompile("[^a-z0-9-]+")
	name = reg.ReplaceAllString(name, "-")
	// Trim hyphens from the start and end
	name = strings.Trim(name, "-")
	return name
}

// truncate shortens a string to a max length, preserving uniqueness where possible.
func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}

// looksLikeEmail determines if a string appears to be intended as an email address
// even if it's malformed (e.g., missing @ symbol)
func looksLikeEmail(s string) bool {
	// If it contains @, it's definitely intended to be an email
	if strings.Contains(s, "@") {
		return true
	}

	// Check for common email domain patterns
	commonDomains := []string{".com", ".org", ".net", ".edu", ".gov", ".co.", ".io", ".dev"}
	for _, domain := range commonDomains {
		if strings.Contains(s, domain) {
			return true
		}
	}

	return false
}

// validateEmail performs basic email validation
// This is copied from your working CLI tool
func validateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// getValidRoles returns a formatted string of valid roles for error messages
func getValidRoles() string {
	roles := make([]string, 0, len(validRoles))
	for role := range validRoles {
		roles = append(roles, role)
	}
	return strings.Join(roles, ", ")
}

// main function sets up the HTTP server and starts listening for requests
func main() {
	// Set up custom HTTP client for skipping SSL verification if needed
	if os.Getenv("NOBL9_SKIP_TLS_VERIFY") == "true" {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		log.Println("WARNING: SSL certificate verification is DISABLED (NOBL9_SKIP_TLS_VERIFY=true)")
	}

	// Register the handler for the create project endpoint
	http.HandleFunc("/api/create-project", handleCreateProject)

	// Get port from environment variable, default to 4000 if not set
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	log.Printf("Go backend listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// handleCreateProject processes HTTP requests to create a new project and assign user roles
func handleCreateProject(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON request body into our struct
	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, false, "Invalid request body: "+err.Error())
		return
	}

	// Validate the request data
	if req.AppID == "" {
		respond(w, false, "Project name (appID) is required")
		return
	}

	if len(req.UserGroups) == 0 {
		respond(w, false, "At least one user group is required")
		return
	}

	// Validate all roles and user identifiers in the request
	for groupIndex, group := range req.UserGroups {
		if !validRoles[group.Role] {
			respond(w, false, fmt.Sprintf("Invalid role '%s' in group %d. Must be one of: %s", group.Role, groupIndex, getValidRoles()))
			return
		}

		// Validate all user identifiers in this group
		userIdentifiers := strings.Split(group.UserIDs, ",")
		for _, userIdentifier := range userIdentifiers {
			userIdentifier = strings.TrimSpace(userIdentifier)
			if userIdentifier == "" {
				continue // Skip empty entries
			}

			// Check if this looks like it's intended to be an email
			if looksLikeEmail(userIdentifier) {
				// This looks like it's intended to be an email, so validate it strictly
				if !validateEmail(userIdentifier) {
					respond(w, false, fmt.Sprintf("Invalid email format: '%s' in group %d. Email addresses must contain @ symbol and be properly formatted (e.g., user@domain.com).", userIdentifier, groupIndex))
					return
				}
			} else {
				// This should be a user ID - validate it's reasonable
				if len(userIdentifier) < 2 {
					respond(w, false, fmt.Sprintf("Invalid user ID: '%s' in group %d (too short)", userIdentifier, groupIndex))
					return
				}
			}
		}
	}

	// Get Nobl9 credentials from environment variables
	clientID := os.Getenv("NOBL9_SDK_CLIENT_ID")
	clientSecret := os.Getenv("NOBL9_SDK_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		respond(w, false, "Missing Nobl9 credentials. Set NOBL9_SDK_CLIENT_ID and NOBL9_SDK_CLIENT_SECRET environment variables.")
		return
	}

	// Create a context with timeout for all SDK operations
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Initialize the Nobl9 client using the same method as your CLI tool
	client, err := sdk.DefaultClient()
	if err != nil {
		respond(w, false, "Failed to initialize Nobl9 SDK client: "+err.Error())
		return
	}

	// Step 1: Check if project already exists
	// We'll try to get the project by creating a project object and checking if it exists
	// Note: The Nobl9 SDK doesn't have a direct "project exists" check, so we'll try to create
	// and handle the error if it already exists

	// Step 2: Create the project
	// Create a new project manifest object with description from the request
	// If no description is provided, use a default one
	description := req.Description
	if description == "" {
		description = fmt.Sprintf("Project created via API: %s", req.AppID)
	}

	project := v1alphaProject.New(
		v1alphaProject.Metadata{
			Name: req.AppID,
		},
		v1alphaProject.Spec{
			Description: description,
		},
	)

	// Step 3: Prepare role bindings for each user group
	var roleBindings []manifest.Object
	var errors []string

	// Process each user group
	for groupIndex, group := range req.UserGroups {
		// Split the comma-separated user IDs/emails
		userIdentifiers := strings.Split(group.UserIDs, ",")

		// Process each user in the group
		for _, userIdentifier := range userIdentifiers {
			userIdentifier = strings.TrimSpace(userIdentifier)
			if userIdentifier == "" {
				continue // Skip empty entries
			}

			// We already validated the format above, so now we just need to process
			var userID string
			if strings.Contains(userIdentifier, "@") {
				// This is an email, try to get the user by email
				log.Printf("Looking up user by email: %s", userIdentifier)
				user, err := client.Users().V2().GetUser(ctx, userIdentifier)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Error retrieving user '%s': %v", userIdentifier, err))
					continue
				}
				if user == nil {
					errors = append(errors, fmt.Sprintf("User with email '%s' not found in Nobl9", userIdentifier))
					continue
				}
				userID = user.UserID
				log.Printf("Found user: %s -> %s", userIdentifier, userID)
			} else {
				// This is a user ID
				userID = userIdentifier
				log.Printf("Using provided user ID: %s", userID)
			}

			// Generate a unique name for the role binding
			// Use the same naming convention as your CLI tool
			sanitizedProject := sanitizeName(req.AppID)
			sanitizedUser := sanitizeName(userIdentifier)
			// Truncate components to ensure the final name is within the 63-char limit required by Nobl9.
			// The name has a fixed overhead: "assign--gX-" + a 10-digit timestamp = ~22 chars.
			// This leaves ~41 chars for the project and user. We'll allocate 20 to each.
			truncatedProject := truncate(sanitizedProject, 20)
			truncatedUser := truncate(sanitizedUser, 20)

			roleBindingName := fmt.Sprintf("assign-%s-%s-g%d-%d",
				truncatedProject,
				truncatedUser,
				groupIndex,
				time.Now().Unix())

			// Create the role binding object
			roleBinding := v1alphaRoleBinding.New(
				v1alphaRoleBinding.Metadata{
					Name: roleBindingName,
				},
				v1alphaRoleBinding.Spec{
					User:       ptr(userID), // Use the user's ID
					RoleRef:    group.Role,  // Role from the request
					ProjectRef: req.AppID,   // Project we just created
				},
			)

			roleBindings = append(roleBindings, roleBinding)
			log.Printf("Created role binding manifest: %s for user %s with role %s", roleBindingName, userID, group.Role)
		}
	}

	// If we had errors finding users, we can't proceed.
	// The project has not been created yet, so we just report the errors.
	if len(errors) > 0 {
		errorMsg := fmt.Sprintf("Failed to create project '%s' because some users could not be found:\n• %s",
			req.AppID, strings.Join(errors, "\n• "))
		respond(w, false, errorMsg)
		return
	}

	// Step 4: Apply the project and all role bindings in a single atomic operation
	allObjects := []manifest.Object{project}
	if len(roleBindings) > 0 {
		allObjects = append(allObjects, roleBindings...)
	}

	if err := client.Objects().V1().Apply(ctx, allObjects); err != nil {
		// Check if the error is because the project already exists
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "conflict") {
			respond(w, false, fmt.Sprintf("Project '%s' already exists", req.AppID))
			return
		}

		respond(w, false, fmt.Sprintf("Failed to create project and assign roles: %v", err))
		return
	}

	log.Printf("Successfully created project '%s' and applied %d role bindings", req.AppID, len(roleBindings))

	// Success! Report back to the client
	message := fmt.Sprintf("Project '%s' created successfully with %d user role assignments", req.AppID, len(roleBindings))
	respond(w, true, message)
}

// respond sends a JSON response to the client
// This helper function makes it easy to send consistent JSON responses
func respond(w http.ResponseWriter, success bool, message string) {
	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Create the response object
	response := Response{
		Success: success,
		Message: message,
	}

	// Encode and send the JSON response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error sending JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	// Log the response for debugging
	if success {
		log.Printf("SUCCESS: %s", message)
	} else {
		log.Printf("ERROR: %s", message)
	}
}
