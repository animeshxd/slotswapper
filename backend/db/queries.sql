-- name: CreateUser :one
INSERT INTO users (
    name,
    email,
    password
) VALUES (
    ?,
    ?,
    ?
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ?;

-- name: GetUserByID :one
SELECT id, name, email, created_at, updated_at FROM users
WHERE id = ?;

-- name: GetPublicUserByID :one
SELECT id, name, created_at, updated_at FROM users
WHERE id = ?;

-- name: CreateEvent :one
INSERT INTO events (
    title,
    start_time,
    end_time,
    status,
    user_id
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING *;

-- name: GetEventByID :one
SELECT * FROM events
WHERE id = ?;

-- name: GetEventsByUserID :many
SELECT * FROM events
WHERE user_id = ?;

-- name: GetEventsByUserIDAndStatus :many
SELECT * FROM events
WHERE user_id = ? AND status = ?;

-- name: UpdateEventStatus :one
UPDATE events
SET status = ?
WHERE id = ?
RETURNING *;

-- name: UpdateEventUserID :one
UPDATE events
SET user_id = ?
WHERE id = ?
RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM events
WHERE id = ?;

-- name: GetSwappableEvents :many
SELECT
    e.id, e.title, e.start_time, e.end_time, e.status, e.user_id, e.created_at, e.updated_at,
    u.name as owner_name
FROM events e
JOIN users u ON e.user_id = u.id
WHERE e.status = 'SWAPPABLE' AND e.user_id != ?;

-- name: CreateSwapRequest :one
INSERT INTO swap_requests (
    requester_user_id,
    responder_user_id,
    requester_slot_id,
    responder_slot_id,
    status
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING *;

-- name: GetSwapRequestByID :one
SELECT * FROM swap_requests
WHERE id = ?;

-- name: GetIncomingSwapRequests :many
SELECT * FROM swap_requests
WHERE responder_user_id = ?;

-- name: GetOutgoingSwapRequests :many
SELECT * FROM swap_requests
WHERE requester_user_id = ?;

-- name: UpdateSwapRequestStatus :one
UPDATE swap_requests
SET status = ?
WHERE id = ?
RETURNING *;
