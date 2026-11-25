-- name: CreateOrganization :one
INSERT INTO organizations (slug, name)
VALUES (@slug, @name)
RETURNING id, slug, name, created_at;

-- name: GetOrganizationBySlug :one
SELECT id, slug, name, created_at
FROM organizations
WHERE slug = @slug;

-- name: ListOrganizations :many
SELECT id, slug, name, created_at
FROM organizations
ORDER BY slug
LIMIT @max_rows;
