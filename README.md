# SlotSwapper


## Design Choices

**Key Design Choices:**

*   **Backend (Go):** Chosen for its performance, concurrency features, and strong typing, which aids in building reliable APIs. `sqlc` is used to generate type-safe Go code from SQL queries, ensuring data integrity and reducing boilerplate. SQLite was used for simplicity in local development.
*   **Frontend (React with TypeScript):** Provides a modern, component-based UI.
    *   TypeScript enhances code quality and maintainability.
    *   `biome` and `tsc` are used for linting and type-checking to catch bugs early in the development cycle.
    *   Libraries like TanStack Router, Zustand, and TanStack Query were used for routing, state management, and data fetching respectively.
*   **UI Components (Shadcn UI):** Utilized for pre-built, accessible, and customizable UI components, accelerating development and ensuring a consistent design.
*   **Authentication (JWT):** Implemented using HTTP-only cookies for secure session management.
*   **Unit/Integration Testing (Go httptest):** Extensive use of Go's `httptest` package for API testing. This approach allows for robust, programmatic verification of backend logic and API responses, reducing reliance on UI-driven testing tools like Postman.

## Running the Application

### With Docker

The easiest way to run the application is with Docker.

1.  **Build the Docker image:**

    ```bash
    docker build -t slotswapper .
    ```

2.  **Run the Docker container:**

    ```bash
    docker run -p 8080:8080 slotswapper
    ```

    The application will be available at [http://localhost:8080](http://localhost:8080).

### Locally

#### Backend

1.  **Navigate to the backend directory:**

    ```bash
    cd backend
    ```

2.  **Install dependencies:**

    ```bash
    go mod tidy
    ```

3.  **Run the backend server:**

    ```bash
    go run ./cmd/slotswapper
    ```

    The backend server will be running on port 8080.

#### Frontend

1.  **Navigate to the frontend directory:**

    ```bash
    cd frontend
    ```

2.  **Install dependencies:**

    ```bash
    pnpm install
    ```

3.  **Run the frontend development server:**

    ```bash
    pnpm dev
    ```

    The frontend will be available at [http://localhost:5173](http://localhost:5173).

## API Endpoints

| Method | Path                                  | Description                                    |
| :----- | :------------------------------------ | :--------------------------------------------- |
| POST   | /api/signup                           | Register a new user.                           |
| POST   | /api/login                            | Log in a user.                                 |
| POST   | /api/logout                           | Log out a user.                                |
| GET    | /api/me                               | Get the current user's profile.                |
| GET    | /api/users/{id}                       | Get a user's public profile.                   |
| POST   | /api/events                           | Create a new event.                            |
| GET    | /api/events/user                      | Get the current user's events.                 |
| GET    | /api/events/{id}                      | Get an event by ID.                            |
| PUT    | /api/events/{id}                      | Update an event.                               |
| POST   | /api/events/{id}/status               | Update an event's status.                      |
| DELETE | /api/events/{id}                      | Delete an event.                               |
| GET    | /api/swappable-slots                  | Get all swappable slots from other users.      |
| POST   | /api/swap-request                     | Create a new swap request.                     |
| GET    | /api/swap-requests/incoming           | Get all incoming swap requests for the user.   |
| GET    | /api/swap-requests/outgoing           | Get all outgoing swap requests from the user.  |
| POST   | /api/swap-response/{id}               | Respond to a swap request.                     |
