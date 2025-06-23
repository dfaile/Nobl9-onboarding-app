# Nobl9 Project Self-Service

A web application for self-service Nobl9 project creation and user role assignment.

## Overview

This web-based self-service tool allows users to create new projects in Nobl9 by specifying an appID (project name) and assigning users or groups with specific roles. The tool streamlines project creation, with a clean webbase user inteface.

## Purpose

- Empower users (including those with read-only permissions) to create and manage Nobl9 projects without admin intervention.
- Ensure projects are easily created and users are assigned appropriate roles.
- Provide clear feedback, validation, for all project creation attempts.
- Built to run inside of docker containers for easy deployment.


## Prerequisites

- Docker and Docker Compose installed
- Nobl9 account and valid Nobl9 API credentials

## Environment Variables

### Backend (Go Service)
The backend requires the following environment variables to be set **at runtime** (not in the Dockerfile):

- `NOBL9_SDK_CLIENT_ID` (your Nobl9 API client ID)
- `NOBL9_SDK_CLIENT_SECRET` (your Nobl9 API client secret)
- `NOBL9_SKIP_TLS_VERIFY` (optional, set to `true` to disable SSL certificate verification for test/dev environments only)

**Set these in your `docker-compose.yml` under the `go-backend` service:**
**Do NOT put secrets in the Dockerfile.**

### Frontend (React App)
- The only relevant environment variable is `REACT_APP_HELP_URL` (optional), which controls the help link in the UI. If not set, it defaults to [https://docs.nobl9.com](https://docs.nobl9.com).
- **Set this variable in your `docker-compose.yml` under the frontend service for local/dev, or in your deployment environment for production.**

## Building and Running

1. **Clone the repository**
   ```bash
   git clone https://github.com/your-org/Nobl9-onboarding-app.git
   cd Nobl9-onboarding-app
   ```
2. **Set your environment variables in `docker-compose.yml`**
3. **Start the stack:**
   ```bash
   docker-compose up --build
   ```
4. **Access the app:**
   - Frontend: [http://localhost](http://localhost)
   - Backend API: [http://localhost:4000](http://localhost:4000)

## API Usage to test (Backend)

### Create Project Endpoint

**Endpoint:** `POST /api/create-project`

**Request Body:**
```json
{
  "appID": "my-project",
  "description": "Optional project description",
  "userGroups": [
    {
      "userIds": "user@example.com,another@example.com",
      "role": "project-owner"
    },
    {
      "userIds": "user123",
      "role": "project-viewer"
    }
  ]
}
```

**Valid Roles:**
- `project-owner`
- `project-editor`
- `project-viewer`

### Example curl
```bash
curl -X POST http://localhost:4000/api/create-project \
  -H "Content-Type: application/json" \
  -d '{
    "appID": "test-project",
    "description": "Test project created via API",
    "userGroups": [
      { "userIds": "user@example.com", "role": "project-owner" },
      { "userIds": "viewer@example.com", "role": "project-viewer" }
    ]
  }'
```

## Frontend Features
- **Project description**: Users can enter a description for the project (optional).
- **Help link**: A "Need help?" link is shown at the bottom of the form, using the `REACT_APP_HELP_URL` variable if set.
- **Role selection**: Only valid roles (`project-owner`, `project-editor`, `project-viewer`) are available in the dropdown.

## Security
- **Never commit secrets** to the repository or Dockerfiles.
- Always set sensitive environment variables in your deployment environment (e.g., `docker-compose.yml`, CI/CD secrets, or Kubernetes manifests).
- **Never set `NOBL9_SKIP_TLS_VERIFY=true` in production!** This disables SSL certificate verification and should only be used for local/test environments with self-signed certificates.

## Setup Instructions

### Local Development
1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd Nobl9-onboarding-app
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start the development server:
   ```bash
   npm start
   ```
   The application will be available at http://localhost:3000.

### Docker Setup
1. Build the Docker image:
   ```bash
   docker build -t nobl9-onboarding-app .
   ```
2. Run the container:
   ```bash
   docker run -p 3000:3000 \
     -e REACT_APP_HELP_URL=https://docs.nobl9.com \
     nobl9-onboarding-app
   ```
   The application will be available at http://localhost:3000.

## Usage Instructions

### Creating a New Project
1. Open the application in your browser.
2. Enter the appID (project name) in the designated field.
   - The appID must contain only letters, numbers, and hyphens.
3. Add user groups by clicking "Add User Group" and entering comma-separated user IDs along with their roles.
   - Each user group must specify a role (Owner, Editor, Viewer, or Integrations user).
   - Maximum 8 users per project.
4. Review the project details in the confirmation dialog.
5. Submit the project. If the project already exists, you will be notified accordingly.

### User Group Management
- You can add up to 8 users per project.
- Each user group must specify a role (Owner, Editor, Viewer, or Integrations user).
- If a user ID does not exist in Nobl9, a configurable help URL will be displayed.

### Admin Notifications
- Admins will be notified of new project creations via email or webhook, as configured.

## Contributing
We welcome contributions! Please follow these steps:
1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Make your changes and ensure tests pass.
4. Submit a pull request with a clear description of your changes.

## Code of Conduct
Please be respectful and inclusive in all interactions. We aim to foster a welcoming and collaborative environment for all contributors.

## Docker Usage
- **Build and Run**: Use `docker build -t nobl9-onboarding-app .` to build the image and `docker run -p 3000:3000 nobl9-onboarding-app` to run it.
- **Docker Compose**: For local development, use `docker-compose up` to start the application with hot-reloading.

## License
This project is licensed under the Mozilla Public License 2.0. See the [LICENSE](LICENSE) file for details. 