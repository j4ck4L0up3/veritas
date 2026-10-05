-- +goose Up
-- content blobs: layers + configs (global, deduped)
CREATE TABLE IF NOT EXISTS blobs (
  digest TEXT PRIMARY KEY,          -- 'sha256:<hex>'
  size INTEGER NOT NULL CHECK (size >= 0),
  created_at INTEGER NOT NULL CHECK (created_at >= 0)          -- unix seconds
);

-- image manifests + indexes (global, deduped). bytes live in blob store
CREATE TABLE IF NOT EXISTS manifests (
  digest TEXT PRIMARY KEY,
  media_type TEXT NOT NULL,             -- manifest.v1+json | index.v1+json
  size INTEGER NOT NULL CHECK (size >= 0),
  created_at INTEGER NOT NULL CHECK (created_at >= 0)
);

-- mutable name -> manifest/index pointer
CREATE TABLE IF NOT EXISTS tags (
  repo TEXT NOT NULL,        -- <name>, e.g. 'myorg/myapp'
  tag_name TEXT NOT NULL,        -- tag, e.g. 'latest'
  manifest_digest TEXT NOT NULL
    REFERENCES manifests(digest) ON DELETE CASCADE,
  updated_at INTEGER NOT NULL CHECK (updated_at >= 0), -- unix seconds
  PRIMARY KEY (repo, tag_name)
);
CREATE INDEX idx_tags_manifest ON tags (manifest_digest);

-- image manifest -> config + layers
CREATE TABLE IF NOT EXISTS manifest_blobs (
  manifest_digest TEXT NOT NULL
    REFERENCES manifests(digest) ON DELETE CASCADE,
  blob_digest TEXT NOT NULL
    REFERENCES blobs(digest) ON DELETE RESTRICT,
  role TEXT NOT NULL CHECK (role IN ('config', 'layer')),
  position INTEGER NOT NULL CHECK (position >= 0),     -- layer order matters
  PRIMARY KEY (manifest_digest, role, position)
);
CREATE INDEX idx_manifest_blobs_blob ON manifest_blobs (blob_digest);

-- index -> child manifests
CREATE TABLE IF NOT EXISTS index_manifests (
  index_digest TEXT NOT NULL
    REFERENCES manifests(digest) ON DELETE CASCADE,
  child_digest TEXT NOT NULL
    REFERENCES manifests(digest) ON DELETE RESTRICT,
  position INTEGER NOT NULL CHECK (position >= 0),
  PRIMARY KEY (index_digest, position)
);
CREATE INDEX idx_index_manifests_child ON index_manifests (child_digest);

-- in-progress blob uploads
CREATE TABLE IF NOT EXISTS upload_sessions (
  uuid TEXT PRIMARY KEY,
  repo TEXT NOT NULL,
  bytes_received INTEGER NOT NULL DEFAULT 0 CHECK (bytes_received >= 0),
  -- unix seconds
  created_at INTEGER NOT NULL CHECK (created_at >= 0),
  -- session invalid after this
  expires_at INTEGER NOT NULL CHECK (expires_at >= 0)
);

-- +goose Down
DROP TABLE upload_sessions;

DROP TABLE index_manifests;

DROP TABLE manifest_blobs;

DROP TABLE blobs;

DROP TABLE manifests;

DROP TABLE tags;
