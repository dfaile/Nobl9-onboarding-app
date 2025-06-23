# Nobl9 Project Creation Service

A Go-based HTTP service that creates Nobl9 projects and assigns user roles using the Nobl9 SDK.

## Prerequisites

- Go 1.16 or higher
- Nobl9 account and credentials
- `curl` for testing

## Environment Variables

Set the following environment variables before running the service:

```bash
export NOBL9_SDK_CLIENT_ID="your-client-id"
export NOBL9_SDK_CLIENT_SECRET="your-client-secret"
export PORT="4000"  # Optional, defaults to 4000
```

## Building

```bash
# Navigate to the service directory
cd cmd/go-backend

# Build the service
go build -o nobl9-project-service

# Run the service
./nobl9-project-service
```

## API Usage

### Create Project Endpoint

**Endpoint:** `POST /api/create-project`

**Request Body:**
```json
{
    "appID": "Test Project",
    "description": "Optional project description",
    "userGroups": [
        {
            "userIds": "user@example.com,another@example.com",
            "role": "project-admin"
        },
        {
            "userIds": "user1@example.com",
            "role": "project-viewer"
        }
    ]
}
```

**Valid Roles:**
- `project-viewer`
- `project-editor`
- `project-owner`

### Testing with curl

```bash
# Create a project with multiple users and roles
curl -X POST http://localhost:4000/api/create-project \
  -H "Content-Type: application/json" \
  -d '{
    "appID": "test-project",
    "description": "Test-Project-From-Curl-'"$(date +%s)"'",,
    "userGroups": [
      {
        "userIds": "editor@example.com",
        "role": "project-editor"
      },
      {
        "userIds": "viewer@example.com",
        "role": "project-viewer"
      }
    ]
  }'

# Create a project with a single user
curl -X POST http://localhost:4000/api/create-project \
  -H "Content-Type: application/json" \
  -d '{
    "appID": "simple-project",
    "userGroups": [
      {
        "userIds": "user@example.com",
        "role": "project-editor"
      }
    ]
  }'
```

## Response Format

```json
{
    "success": true,
    "message": "Project 'test-project' created successfully with 2 user role assignments"
}
```

## Error Handling

The service validates:
- Project name (appID) is required
- At least one user group is required
- Valid role assignments
- Email format for email addresses
- User ID format for non-email identifiers

Error responses will include detailed messages about what went wrong.

## Notes

- The service automatically sanitizes project and user names to be RFC-1123 compliant
- User identifiers can be either email addresses or user IDs
- The service will attempt to look up users by email if an email address is provided
- If user assignment fails, the project will still be created but an error will be returned 