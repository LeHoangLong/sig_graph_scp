CREATE TABLE IF NOT EXISTS gorm_request_to_accept_assets (
    id BIGSERIAL PRIMARY KEY,
    request_status VARCHAR(256) NOT NULL,
    is_outbound_or_inbound BOOLEAN NOT NULL,
    request_time_ms BIGINT NOT NULL,
    ack_id VARCHAR(1024) NOT NULL,
    asset_id BIGINT NOT NULL REFERENCES gorm_nodes(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    new_asset_id BIGINT REFERENCES gorm_nodes(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    peer_id BIGINT NOT NULL REFERENCES gorm_peers(id) ON DELETE CASCADE ON UPDATE CASCADE,
    user_id BIGINT NOT NULL,
    accept_message VARCHAR(8192) NOT NULL
);

CREATE TABLE IF NOT EXISTS gorm_request_to_accept_asset_exposed_private_ids (
    id BIGSERIAL PRIMARY KEY,
    request_id BIGINT REFERENCES gorm_request_to_accept_assets(id) ON DELETE CASCADE ON UPDATE CASCADE,

    this_hash VARCHAR(1024) NOT NULL, 
    this_id VARCHAR(1024) NOT NULL,
    this_secret VARCHAR(1024) NOT NULL DEFAULT '',

    other_hash VARCHAR(1024) NOT NULL, 
    other_id VARCHAR(1024) NOT NULL,
    other_secret VARCHAR(1024) NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS gorm_request_to_accept_asset_candidate_ids (
    id BIGSERIAL PRIMARY KEY,
    request_id BIGINT REFERENCES gorm_request_to_accept_assets(id) ON DELETE CASCADE ON UPDATE CASCADE,

    candidate_id VARCHAR(1024) NOT NULL, 
    candidate_secret VARCHAR(1024) NOT NULL,
    candidate_signature VARCHAR(1024) NOT NULL DEFAULT ''
);
