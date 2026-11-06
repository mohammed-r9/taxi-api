-- name: GetRolePermissions :many
SELECT p.name
FROM permissions p
JOIN roles_permissions rp ON rp.permission_id = p.id
JOIN roles r ON r.id = rp.role_id
WHERE rp.role_id = @role_id
  AND p.deleted_at IS NULL
  AND r.deleted_at IS NULL
ORDER BY p.name;
