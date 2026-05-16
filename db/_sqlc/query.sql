-- name: addPipeline :exec
WITH ins AS (
  INSERT INTO pipeline (id, created_at) 
  VALUES ($1, $2)
  RETURNING id
) INSERT INTO pipeline_history (
  pipeline_id, updated_at, nodes_per_stage, nodes, version
) SELECT $1, $2, $3, $4, 1 FROM ins;

-- name: GetPipelineHistory :many
SELECT * FROM pipeline_history
WHERE pipeline_id = $1
ORDER BY version DESC;

-- name: GetLatestPipeline :one
SELECT * FROM pipeline_history
WHERE pipeline_id = $1
ORDER BY VERSION DESC
LIMIT 1;

-- name: GetLatestNode :one
SELECT n.id, n.type, n.created_at FROM node n
JOIN LATERAL (
  SELECT updated_at, version, body FROM node_history nh
  WHERE nh.node_id = n.id
  ORDER BY nh.version DESC LIMIT 1
) on TRUE
WHERE n.id = $1;

-- name: GetLatestNodes :many
SELECT n.id, n.type, n.created_at FROM node n
JOIN LATERAL (
  SELECT updated_at, version, body FROM node_history nh
  WHERE nh.node_id = n.id
  ORDER BY nh.version DESC LIMIT 1
) on TRUE
WHERE n.id = ANY(@ids::uuid[]);

-- name: GetNodeHistory :one
SELECT * FROM node_history
WHERE node_id = $1
ORDER BY version DESC;

-- name: GetNodesHistory :many
SELECT * FROM node_history
WHERE node_id = ANY(@ids::uuid[])
ORDER BY version DESC;
