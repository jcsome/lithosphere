DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'uint256') THEN
        CREATE DOMAIN UINT256 AS NUMERIC
            CHECK (VALUE >= 0 AND VALUE < POWER(CAST(2 AS NUMERIC), CAST(256 AS NUMERIC)) AND SCALE(VALUE) = 0);
    ELSE
        ALTER DOMAIN UINT256 DROP CONSTRAINT uint256_check;
        ALTER DOMAIN UINT256 ADD
            CHECK (VALUE >= 0 AND VALUE < POWER(CAST(2 AS NUMERIC), CAST(256 AS NUMERIC)) AND SCALE(VALUE) = 0);
    END IF;
END $$;


CREATE TABLE IF NOT EXISTS l1_block_headers (
    hash        VARCHAR PRIMARY KEY,
    parent_hash VARCHAR NOT NULL UNIQUE,
    number      UINT256 NOT NULL UNIQUE,
    timestamp   INTEGER NOT NULL UNIQUE CHECK (timestamp > 0),
    rlp_bytes   VARCHAR NOT NULL
);
CREATE INDEX IF NOT EXISTS l1_block_headers_timestamp ON l1_block_headers(timestamp);
CREATE INDEX IF NOT EXISTS l1_block_headers_number ON l1_block_headers(number);

CREATE TABLE IF NOT EXISTS l2_block_headers (
    hash        VARCHAR PRIMARY KEY,
    parent_hash VARCHAR NOT NULL UNIQUE,
    number      UINT256 NOT NULL UNIQUE,
    timestamp   INTEGER NOT NULL,
    rlp_bytes   VARCHAR NOT NULL
);
CREATE INDEX IF NOT EXISTS l2_block_headers_timestamp ON l2_block_headers(timestamp);
CREATE INDEX IF NOT EXISTS l2_block_headers_number ON l2_block_headers(number);

CREATE TABLE IF NOT EXISTS l1_contract_events (
    guid             VARCHAR PRIMARY KEY,
    block_hash       VARCHAR NOT NULL REFERENCES l1_block_headers(hash) ON DELETE CASCADE,
    contract_address VARCHAR NOT NULL,
    transaction_hash VARCHAR NOT NULL,
    log_index        INTEGER NOT NULL,
    event_signature  VARCHAR NOT NULL,
    timestamp        INTEGER NOT NULL CHECK (timestamp > 0),
    rlp_bytes        VARCHAR NOT NULL
);
CREATE INDEX IF NOT EXISTS l1_contract_events_timestamp ON l1_contract_events(timestamp);
CREATE INDEX IF NOT EXISTS l1_contract_events_block_hash ON l1_contract_events(block_hash);
CREATE INDEX IF NOT EXISTS l1_contract_events_event_signature ON l1_contract_events(event_signature);
CREATE INDEX IF NOT EXISTS l1_contract_events_contract_address ON l1_contract_events(contract_address);

CREATE TABLE IF NOT EXISTS l2_contract_events (
    -- Searchable fields
    guid             VARCHAR PRIMARY KEY,
    block_hash       VARCHAR NOT NULL REFERENCES l2_block_headers(hash) ON DELETE CASCADE,
    contract_address VARCHAR NOT NULL,
    transaction_hash VARCHAR NOT NULL,
    log_index        INTEGER NOT NULL,
    event_signature  VARCHAR NOT NULL, -- bytes32(0x0) when topics are missing
    timestamp        INTEGER NOT NULL CHECK (timestamp > 0),

    -- Raw Data
    rlp_bytes VARCHAR NOT NULL
);
CREATE INDEX IF NOT EXISTS l2_contract_events_timestamp ON l2_contract_events(timestamp);
CREATE INDEX IF NOT EXISTS l2_contract_events_block_hash ON l2_contract_events(block_hash);
CREATE INDEX IF NOT EXISTS l2_contract_events_event_signature ON l2_contract_events(event_signature);
CREATE INDEX IF NOT EXISTS l2_contract_events_contract_address ON l2_contract_events(contract_address);


CREATE TABLE IF NOT EXISTS l1_to_l2 (
    guid                    VARCHAR PRIMARY KEY,
    l1_block_number         UINT256 NOT NULL,
    l2_block_number         UINT256,
    queue_index             UINT256,
    l1_transaction_hash     VARCHAR NOT NULL,
    l2_transaction_hash     VARCHAR NOT NULL,
    transaction_source_hash VARCHAR NOT NULL,
    message_hash            VARCHAR,
    l1_tx_origin            VARCHAR,
    status                  SMALlINT NOT NULL,
    from_address            VARCHAR NOT NULL,
    to_address              VARCHAR NOT NULL,
    l1_token_address        VARCHAR,
    l2_token_address        VARCHAR,
    eth_amount              UINT256,
    erc20_amount            UINT256,
    gas_limit               UINT256 NOT NULL,
    version                 INTEGER DEFAULT 0,
    timestamp               INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS l1_to_l2_timestamp ON l1_to_l2(timestamp);
CREATE INDEX IF NOT EXISTS l1_to_l2_l1_transaction_hash ON l1_to_l2(l1_transaction_hash);
CREATE INDEX IF NOT EXISTS l1_to_l2_l2_transaction_hash ON l1_to_l2(l2_transaction_hash);
CREATE INDEX IF NOT EXISTS l1_to_l2_from_address ON l1_to_l2(from_address);
CREATE INDEX IF NOT EXISTS l1_to_l2_to_address ON l1_to_l2(to_address);
CREATE INDEX IF NOT EXISTS l1_to_l2_message_hash ON l1_to_l2(message_hash);
CREATE INDEX IF NOT EXISTS l1_to_l2_transaction_source_hash ON l1_to_l2(transaction_source_hash);

CREATE TABLE IF NOT EXISTS withdraw_proven (
    guid                          VARCHAR PRIMARY KEY,
    block_number                  UINT256 NOT NULL,
    withdraw_hash                 VARCHAR NOT NULL,
    message_hash                  VARCHAR,
    proven_transaction_hash       VARCHAR NOT NULL,
    l1_token_address              VARCHAR,
    l2_token_address              VARCHAR,
    eth_amount                    UINT256,
    erc20_amount                  UINT256,
    related                       BOOLEAN DEFAULT FALSE,
    timestamp                     INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS withdraw_proven_withdrawal_hash ON withdraw_proven(withdraw_hash);
CREATE INDEX IF NOT EXISTS withdraw_proven_timestamp ON withdraw_proven(timestamp);

CREATE TABLE IF NOT EXISTS withdraw_finalized (
    guid                          VARCHAR PRIMARY KEY,
    block_number                  UINT256 NOT NULL,
    withdraw_hash                 VARCHAR NOT NULL,
    message_hash                  VARCHAR,
    finalized_transaction_hash    VARCHAR NOT NULL,
    l1_token_address              VARCHAR,
    l2_token_address              VARCHAR,
    eth_amount                    UINT256,
    erc20_amount                  UINT256,
    related                       BOOLEAN DEFAULT FALSE,
    timestamp                     INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS withdraw_finalized_withdrawal_hash ON withdraw_finalized(withdraw_hash);
CREATE INDEX IF NOT EXISTS withdraw_finalized_timestamp ON withdraw_finalized(timestamp);

CREATE TABLE IF NOT EXISTS l2_to_l1 (
    guid                         VARCHAR PRIMARY KEY,
    l1_block_number              UINT256,
    l2_block_number              UINT256 NOT NULL,
    msg_nonce                    UINT256 NOT NULL,
    l2_transaction_hash          VARCHAR NOT NULL,
    withdraw_transaction_hash    VARCHAR NOT NULL,
    message_hash                 VARCHAR,
    l1_prove_tx_hash             VARCHAR,
    l1_finalize_tx_hash          VARCHAR,
    status                       SMALLINT NOT NULL,
    from_address                 VARCHAR NOT NULL,
    to_address                   VARCHAR NOT NULL,
    l1_token_address             VARCHAR,
    l2_token_address             VARCHAR,
    eth_amount                   UINT256,
    erc20_amount                 UINT256,
    gas_limit                    UINT256 NOT NULL,
    time_left                    UINT256 NOT NULL,
    version                      INTEGER DEFAULT 0,
    timestamp                    INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS l2_to_l1_timestamp ON l2_to_l1(timestamp);
CREATE INDEX IF NOT EXISTS l2_to_l1_l2_transaction_hash ON l2_to_l1(l2_transaction_hash);
CREATE INDEX IF NOT EXISTS l2_to_l1_l1_prove_tx_hash ON l2_to_l1(l1_prove_tx_hash);
CREATE INDEX IF NOT EXISTS l2_to_l1_l1_finalize_tx_hash ON l2_to_l1(l1_finalize_tx_hash);
CREATE INDEX IF NOT EXISTS l2_to_l1_from_address ON l2_to_l1(from_address);
CREATE INDEX IF NOT EXISTS l2_to_l1_to_address ON l2_to_l1(to_address);
CREATE INDEX IF NOT EXISTS l2_to_l1_message_hash ON l2_to_l1(message_hash);
CREATE INDEX IF NOT EXISTS l2_to_l1_withdraw_transaction_hash ON l2_to_l1(withdraw_transaction_hash);

CREATE TABLE IF NOT EXISTS state_root (
    guid                      VARCHAR PRIMARY KEY,
    block_hash                VARCHAR NOT NULL,
    transaction_hash          VARCHAR NOT NULL,
    l1_block_number           UINT256 DEFAULT 0,
    l2_block_number           UINT256 DEFAULT 0,
    output_index              UINT256 NOT NULL,
    prev_total_elements       UINT256 DEFAULT 0,
    status                    SMALLINT NOT NULL default 0,
    output_root               VARCHAR NOT NULL,
    canonical                 BOOLEAN DEFAULT TRUE,
    batch_size                UINT256 NOT NULL,
    block_size                UINT256 NOT NULL,
    timestamp                 INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS state_root_block_hash ON state_root(block_hash);
CREATE INDEX IF NOT EXISTS state_root_transaction_hash ON state_root(transaction_hash);

CREATE TABLE IF NOT EXISTS data_store_event (
    guid                 VARCHAR PRIMARY KEY,
    data_store_id        INTEGER NOT NULL,
    block_hash           VARCHAR NOT NULL,
    timestamp            INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS data_store_event_data_store_id ON data_store_event(data_store_id);
CREATE INDEX IF NOT EXISTS data_store_event_timestamp ON data_store_event(timestamp);


CREATE TABLE IF NOT EXISTS data_store (
    guid                    VARCHAR PRIMARY KEY,
    data_store_id           UINT256 NOT NULL,
    duration_data_store_id  UINT256 NOT NULL CHECK (duration_data_store_id > 0),
    data_size               UINT256 NOT NULL,
    data_init_hash          VARCHAR NOT NULL UNIQUE REFERENCES data_store(data_init_hash) ON DELETE CASCADE,
    data_confirm_hash       VARCHAR NOT NULL UNIQUE REFERENCES data_store(data_confirm_hash) ON DELETE CASCADE,
    from_store_number       UINT256 NOT NULL,
    stake_from_block_number UINT256 NOT NULL,
    init_gas_used           UINT256 NOT NULL,
    init_block_number       UINT256 NOT NULL,
    confirm_gas_used        UINT256 NOT NULL,
    data_hash               VARCHAR NOT NULL,
    eth_sign                VARCHAR NOT NULL,
    mantle_sign             VARCHAR NOT NULL,
    signatory_record        VARCHAR NOT NULL,
    init_time               INTEGER NOT NULL CHECK (init_time > 0),
    expire_time             INTEGER NOT NULL CHECK (expire_time > 0),
    num_sys                 SMALLINT NOT NULL,
    num_par                 SMALLINT NOT NULL,
    low_degree              VARCHAR NOT NULL,
    status                  BOOLEAN DEFAULT FALSE,
    confirmer               VARCHAR NOT NULL,
    header                  VARCHAR NOT NULL,
    data_commitment         VARCHAR NOT NULL,
    timestamp               INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS data_store_timestamp ON data_store(timestamp);
CREATE INDEX IF NOT EXISTS data_store_duration_data_store_id ON data_store(duration_data_store_id);
CREATE INDEX IF NOT EXISTS data_store_data_store_id ON data_store(data_store_id);
CREATE INDEX IF NOT EXISTS data_store_data_hash ON data_store(data_hash);


CREATE TABLE IF NOT EXISTS data_store_block (
    guid                 VARCHAR PRIMARY KEY,
    data_store_id        INTEGER NOT NULL,
    block_data           VARCHAR NOT NULL,
    l2_transaction_hash  VARCHAR NOT NULL,
    l2_block_number      UINT256,
    canonical            BOOLEAN DEFAULT TRUE,
    timestamp            INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS data_store_block_data_store_id ON data_store_block(data_store_id);
CREATE INDEX IF NOT EXISTS data_store_block_timestamp ON data_store_block(timestamp);


CREATE TABLE IF NOT EXISTS transactions (
    guid                      VARCHAR PRIMARY KEY,
    block_hash                VARCHAR NOT NULL,
    block_number              UINT256 DEFAULT 0,
    from_address              VARCHAR NOT NULL,
    to_address                VARCHAR,
    gas                       UINT256 NOT NULL,
    gas_price                 UINT256 NOT NULL,
    hash                      VARCHAR NOT NULL,
    input_data                VARCHAR,
    max_fee_per_gas           UINT256 DEFAULT 0,
    max_priority_fee_per_gas  UINT256 DEFAULT 0,
    gas_used                  UINT256 DEFAULT 0,
    cumulative_gas_used       UINT256 DEFAULT 0,
    effective_gas_price       UINT256 DEFAULT 0,
    l1_fee                    UINT256 DEFAULT 0,
    L1_gas_used               UINT256 DEFAULT 0,
    l1_gas_price              UINT256 DEFAULT 0,
    nonce                     UINT256,
    transaction_index         UINT256 NOT NULL,
    tx_type                   SMALLINT NOT NULL default 0,
    r                         VARCHAR NOT NULL,
    s                         VARCHAR NOT NULL,
    v                         VARCHAR NOT NULL,
    status                    SMALLINT NOT NULL default 0,
    contract_address          VARCHAR,
    amount                    UINT256 DEFAULT 0,
    y_parity                  VARCHAR NOT NULL,
    timestamp                 INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS transactions_block_number ON transactions(block_number);
CREATE INDEX IF NOT EXISTS transactions_hash ON transactions(hash);
CREATE INDEX IF NOT EXISTS transactions_timestamp ON transactions(timestamp);


