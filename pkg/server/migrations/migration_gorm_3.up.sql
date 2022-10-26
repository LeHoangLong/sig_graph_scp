CREATE TABLE IF NOT EXISTS gorm_request_to_accept_assets (
    id BIGSERIAL PRIMARY KEY,
    request_status VARCHAR(256) NOT NULL,
    is_outbound_or_inbound BOOLEAN NOT NULL,
    request_time_ms BIGINT NOT NULL,
    ack_id VARCHAR(1024) NOT NULL,
    accepted BOOLEAN NOT NULL,
    asset_id BIGINT NOT NULL REFERENCES gorm_nodes(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    peer_id BIGINT NOT NULL REFERENCES gorm_peers(id) ON DELETE CASCADE ON UPDATE CASCADE,
    user_id BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS gorm_request_to_accept_asset_exposed_private_ids (
    id BIGSERIAL PRIMARY KEY,
    request_id BIGINT REFERENCES gorm_request_to_accept_assets(id) ON DELETE CASCADE ON UPDATE CASCADE,

    this_hash VARCHAR(1024) NOT NULL, 
    this_node_id VARCHAR(1024),
    this_secret VARCHAR(1024) NOT NULL DEFAULT '',

    other_hash VARCHAR(1024) NOT NULL, 
    other_node_id VARCHAR(1024),
    other_secret VARCHAR(1024) NOT NULL DEFAULT ''
);