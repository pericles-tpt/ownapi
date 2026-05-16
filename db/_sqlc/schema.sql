CREATE TABLE pipeline (
  id UUID PRIMARY KEY NOT NULL,
  created_at TIMESTAMP NOT NULL
);

CREATE TABLE pipeline_draft (
  id UUID PRIMARY KEY NOT NULL,
  created_at TIMESTAMP NOT NULL
);

CREATE TABLE pipeline_history (
  id SERIAL PRIMARY KEY,
  pipeline_id UUID NOT NULL,
  nodes_per_stage INTEGER[],
  updated_at TIMESTAMP NOT NULL,
  nodes UUID[],
  node_versions INT[],
  version INT
);

CREATE TABLE node (
  id UUID PRIMARY KEY,
  type INTEGER,
  created_at TIMESTAMP NOT NULL
);

CREATE TABLE node_draft (
  id UUID PRIMARY KEY,
  type INTEGER,
  created_at TIMESTAMP NOT NULL
);

CREATE TABLE node_history (
  id SERIAL PRIMARY KEY,
  node_id UUID,
  updated_at TIMESTAMP NOT NULL,
  version INT,
  body JSONB
);

CREATE TABLE node_aliases (
  id SERIAL PRIMARY KEY,
  node_id uuid,
  alias TEXT
)
