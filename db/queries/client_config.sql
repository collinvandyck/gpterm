-- name: GetClientConfig :one
SELECT * FROM client_config
where name = (select value from config where name = 'client-config');

-- name: UpdateClientConfig :one
update client_config
set message_context = ?
where name = (select value from config where name = 'client-config')
returning *;

-- name: CycleClientConfig :one
UPDATE config
SET value = (
    SELECT
        CASE
            -- Check if current value is less than the maximum value
            WHEN (SELECT value FROM config WHERE name = 'client-config') < (SELECT MAX(name) FROM client_config)
            THEN (
                -- If current value is less than the maximum value, select the next value
                SELECT MIN(name) FROM client_config
                WHERE name > (SELECT value FROM config WHERE name = 'client-config')
            )
            ELSE (
                -- If current value is equal to the maximum value, wrap around to the beginning
                SELECT MIN(name) FROM client_config
            )
        END
)
WHERE name = 'client-config'
RETURNING *;

