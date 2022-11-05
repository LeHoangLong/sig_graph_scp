CREATE TABLE IF NOT EXISTS gorm_peer_protocols (
    id SERIAL PRIMARY KEY,
    protocol_type VARCHAR(256) NOT NULL,
    version_major INTEGER NOT NULL,
    version_minor INTEGER NOT NULL 
);

CREATE TABLE IF NOT EXISTS gorm_peers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    connection_uri VARCHAR(4096) NOT NULL,
    peer_pem_public_key VARCHAR(4096) NOT NULL,
    protocol_id INTEGER REFERENCES gorm_peer_protocols(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

INSERT INTO gorm_peer_protocols (
    protocol_type,
    version_major,
    version_minor
) VALUES (
    'grpc',
    1,
    0
);
