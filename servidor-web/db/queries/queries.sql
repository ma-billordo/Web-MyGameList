-- USERS
-- name: CreateUser :one
INSERT INTO Users (name, email, password)
VALUES ($1, $2, $3)
RETURNING id_user;

-- name: GetUser :one
SELECT id_user, name, email
FROM Users
WHERE id_user = $1;

-- name: ListUsers :many
SELECT id_user, name, email
FROM Users
ORDER BY id_user;

-- name: UpdateUser :exec
UPDATE Users
SET name = $2, email =$3
WHERE id_user = $1;

-- name: DeleteUser :exec
DELETE FROM Users
WHERE id_user = $1;

-- LIBRARY
-- name: CreateGame :one
INSERT INTO Library (title, description)
VALUES ($1, $2)
RETURNING id_game;

-- name: GetGame :one
SELECT id_game, title, description, avg_score
FROM Library
WHERE id_game = $1;

-- name: ListGames :many
SELECT id_game, title, description, avg_score
FROM Library
ORDER BY id_game;

-- name: UpdateGame :exec
UPDATE Library
SET title = $2, description =$3
WHERE id_game = $1;

-- name: DeleteGame :exec
DELETE FROM Library
WHERE id_game = $1;

-- CATALOGUE
-- name: CreateRecord :exec
INSERT INTO Catalogue (id_user, id_game, score, status)
VALUES ($1, $2, $3, $4);

-- name: GetRecord :one
SELECT id_user, id_game, score, status
FROM Catalogue
WHERE id_user = $1 AND id_game = $2;

-- name: LisRecord :many
SELECT id_user, id_game, score, status
FROM Catalogue
WHERE id_user = $1
ORDER BY (id_user, id_game);

-- name: UpdateRecord :exec
UPDATE Catalogue
SET score = $3, status =$4
WHERE id_user = $1 AND id_game = $2;

-- name: DeleteRecord :exec
DELETE FROM Catalogue
WHERE id_user = $1 AND id_game = $2;
