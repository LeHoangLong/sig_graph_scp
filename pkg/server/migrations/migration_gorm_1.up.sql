CREATE TABLE IF NOT EXISTS gorm_nodes (
    id BIGSERIAL PRIMARY KEY,
    node_id VARCHAR(1024) NOT NULL,
    node_namespace VARCHAR(1024) NOT NULL,
    node_type VARCHAR(128) NOT NULL,
    is_finalized BOOLEAN NOT NULL,
    created_time BIGINT NOT NULL,
    updated_time BIGINT NOT NULL,
    node_signature VARCHAR(1024) NOT NULL,
    owner_public_key VARCHAR(4096) NOT NULL,
    UNIQUE(node_id, node_namespace)
);

CREATE TABLE IF NOT EXISTS gorm_public_edges (
    node_db_id BIGINT REFERENCES gorm_nodes(id) ON UPDATE CASCADE ON DELETE CASCADE,
    other_node_id VARCHAR(1024) NOT NULL,
    PRIMARY KEY(node_db_id, other_node_id)
);

CREATE TABLE IF NOT EXISTS gorm_private_edges (
    node_db_id BIGINT REFERENCES gorm_nodes(id) ON UPDATE CASCADE ON DELETE CASCADE,
    other_node_hash VARCHAR(1024) NOT NULL, 
    other_node_id VARCHAR(1024),
    other_node_id_secret VARCHAR(1024) NOT NULL DEFAULT '',
    PRIMARY KEY(node_db_id, other_node_hash)
);

CREATE TABLE IF NOT EXISTS gorm_assets (
    node_db_id BIGINT REFERENCES gorm_nodes(id) ON UPDATE CASCADE ON DELETE CASCADE PRIMARY KEY,
    creation_process VARCHAR(16) NOT NULL,
    unit VARCHAR(1024) NOT NULL,
    quantity NUMERIC(32, 16) NOT NULL,
    material_name VARCHAR(1024) NOT NULL
);

CREATE TABLE IF NOT EXISTS gorm_user_key_pairs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    public_key  VARCHAR(4096) NOT NULL UNIQUE,
    private_key  VARCHAR(4096) NOT NULL
);
