-- name: CreateSecret :execresult
-- Inserts a new secret into the table and returns the ID of the newly inserted row
INSERT INTO secrets (
    token,
    text,
    password,
    expiration_date
) VALUES (
    ?,
    ?,
    ?,
    ?
);

-- name: GetSecretByToken :one
-- Retrieves a single secret from the table based on the token
SELECT * FROM secrets
WHERE token = ?;

-- name: MarkSecretAsBurned :execresult
-- Marks a secret as burned by setting is_burned to TRUE for the given token
UPDATE secrets
SET is_burned = TRUE
WHERE token = ?;

-- name: ListAllSecrets :many
-- Lists all secrets in the table
SELECT * FROM secrets;

-- name: DeleteSecretByToken :execresult
-- Deletes a secret from the table based on the token
DELETE FROM secrets
WHERE token = ?;

-- name: CheckSecretExpiration :one
-- Checks if a secret with the given token has expired
SELECT COUNT(*) > 0 AS has_been_purged
FROM secrets
WHERE token = ? AND expiration_date < NOW();

-- name: MarkSecretAsViewed :execresult
-- Marks a secret as viewed by setting is_viewed to TRUE for the given token
UPDATE secrets
SET is_viewed = TRUE
WHERE token = ?;
