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

-- name: UpdateEvent :one
UPDATE events
SET title = ?,
    start_time = ?,
    end_time = ?,
    status = ?
WHERE id = ?
RETURNING *;

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

-- name: UpdateSwapRequestStatus :one
UPDATE swap_requests
SET status = ?
WHERE id = ?
RETURNING *;

-- name: DeleteSwapRequest :exec
DELETE FROM swap_requests
WHERE id = ?;

-- name: GetSwapRequestsByEventID :many
SELECT * FROM swap_requests
WHERE requester_slot_id = ? OR responder_slot_id = ?;

-- name: GetIncomingSwapRequests :many
SELECT
    sr.id,
    sr.status,
    sr.requester_user_id,
    requester.name AS requester_name,
    requester_event.title AS requester_event_title,
    requester_event.start_time AS requester_event_start_time,
    requester_event.end_time AS requester_event_end_time,
    responder_event.title AS responder_event_title,
    responder_event.start_time AS responder_event_start_time,
    responder_event.end_time AS responder_event_end_time
FROM
    swap_requests sr
JOIN
    users requester ON sr.requester_user_id = requester.id
JOIN
    events requester_event ON sr.requester_slot_id = requester_event.id
JOIN
    events responder_event ON sr.responder_slot_id = responder_event.id
WHERE
    sr.responder_user_id = ? AND sr.status = 'PENDING';

-- name: GetOutgoingSwapRequests :many
SELECT
    sr.id,
    sr.status,
    sr.responder_user_id,
    responder.name AS responder_name,
    requester_event.title AS requester_event_title,
    requester_event.start_time AS requester_event_start_time,
    requester_event.end_time AS requester_event_end_time,
    responder_event.title AS responder_event_title,
    responder_event.start_time AS responder_event_start_time,
    responder_event.end_time AS responder_event_end_time
FROM
    swap_requests sr
JOIN
    users responder ON sr.responder_user_id = responder.id
JOIN
    events requester_event ON sr.requester_slot_id = requester_event.id
JOIN
    events responder_event ON sr.responder_slot_id = responder_event.id
WHERE
    sr.requester_user_id = ? AND sr.status = 'PENDING';