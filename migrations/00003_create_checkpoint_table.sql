CREATE TABLE IF NOT EXISTS bridge_checkpoints (
    id SERIAL PRIMARY KEY,
    snapshot_time  timestamp,
    l1_number UINT256,
    l1_token_address VARCHAR,
    l2_number UINT256,
    l2_token_address VARCHAR,
    l1_bridge_balance UINT256,
    total_supply VARCHAR,
    checked BOOLEAN,
    status SMALLINT NOT NULL
);

CREATE TABLE IF NOT EXISTS daily_stat (
    guid                    VARCHAR PRIMARY KEY,
    tx_count                UINT256 DEFAULT 0,
    active_user             UINT256 DEFAULT 0,
    nft_holders             UINT256 DEFAULT 0,
    new_user                UINT256 DEFAULT 0,
    deposit_count           UINT256 DEFAULT 0,
    withdraw_count          UINT256 DEFAULT 0,
    deposit_amount          UINT256 DEFAULT 0,
    withdraw_amount         UINT256 DEFAULT 0,
    developer_count         UINT256 DEFAULT 0,
    smart_contract_count    UINT256 DEFAULT 0,
    l1_cost_amount          UINT256 DEFAULT 0,
    l2_fee_amount           UINT256 DEFAULT 0,
    timestamp               INTEGER NOT NULL CHECK (timestamp > 0)
 );
CREATE INDEX IF NOT EXISTS daily_stat_timestamp ON daily_stat(timestamp);


CREATE TABLE IF NOT EXISTS weekly_stat (
    guid                    VARCHAR PRIMARY KEY,
    tx_count                UINT256 DEFAULT 0,
    active_user             UINT256 DEFAULT 0,
    nft_holders             UINT256 DEFAULT 0,
    new_user                UINT256 DEFAULT 0,
    deposit_count           UINT256 DEFAULT 0,
    withdraw_count          UINT256 DEFAULT 0,
    deposit_amount          UINT256 DEFAULT 0,
    withdraw_amount         UINT256 DEFAULT 0,
    developer_count         UINT256 DEFAULT 0,
    smart_contract_count    UINT256 DEFAULT 0,
    l1_cost_amount          UINT256 DEFAULT 0,
    l2_fee_amount           UINT256 DEFAULT 0,
    timestamp               INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS weekly_stat_timestamp ON weekly_stat(timestamp);


CREATE TABLE IF NOT EXISTS monthly_stat (
    guid                    VARCHAR PRIMARY KEY,
    tx_count                UINT256 DEFAULT 0,
    active_user             UINT256 DEFAULT 0,
    nft_holders             UINT256 DEFAULT 0,
    new_user                UINT256 DEFAULT 0,
    deposit_count           UINT256 DEFAULT 0,
    withdraw_count          UINT256 DEFAULT 0,
    deposit_amount          UINT256 DEFAULT 0,
    withdraw_amount         UINT256 DEFAULT 0,
    developer_count         UINT256 DEFAULT 0,
    smart_contract_count    UINT256 DEFAULT 0,
    l1_cost_amount          UINT256 DEFAULT 0,
    l2_fee_amount           UINT256 DEFAULT 0,
    timestamp               INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS monthly_stat_timestamp ON monthly_stat(timestamp);


CREATE TABLE IF NOT EXISTS cumulative_stat (
    guid                    VARCHAR PRIMARY KEY,
    tx_count                UINT256 DEFAULT 0,
    active_user             UINT256 DEFAULT 0,
    nft_holders             UINT256 DEFAULT 0,
    new_user                UINT256 DEFAULT 0,
    deposit_count           UINT256 DEFAULT 0,
    withdraw_count          UINT256 DEFAULT 0,
    deposit_amount          UINT256 DEFAULT 0,
    withdraw_amount         UINT256 DEFAULT 0,
    developer_count         UINT256 DEFAULT 0,
    smart_contract_count    UINT256 DEFAULT 0,
    l1_cost_amount          UINT256 DEFAULT 0,
    l2_fee_amount           UINT256 DEFAULT 0,
    timestamp               INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS cumulative_stat_timestamp ON cumulative_stat(timestamp);


CREATE TABLE IF NOT EXISTS symbol_daily_tvl (
    guid                VARCHAR PRIMARY KEY,
    symbol              VARCHAR NOT NULL,
    amount              UINT256 DEFAULT 0,
    latest_price        VARCHAR,
    trans_to_usd        UINT256 DEFAULT 0,
    timestamp           INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS symbol_daily_tvl_timestamp ON symbol_daily_tvl(timestamp);


CREATE TABLE IF NOT EXISTS symbol_weekly_tvl (
    guid                VARCHAR PRIMARY KEY,
    symbol              VARCHAR NOT NULL,
    amount              UINT256 DEFAULT 0,
    latest_price        VARCHAR,
    trans_to_usd        UINT256 DEFAULT 0,
    timestamp           INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS symbol_weekly_tvl_timestamp ON symbol_weekly_tvl(timestamp);


CREATE TABLE IF NOT EXISTS symbol_monthly_tvl (
     guid                VARCHAR PRIMARY KEY,
     symbol              VARCHAR NOT NULL,
     amount              UINT256 DEFAULT 0,
     latest_price        VARCHAR,
     trans_to_usd        UINT256 DEFAULT 0,
     timestamp           INTEGER NOT NULL CHECK (timestamp > 0)
 );
CREATE INDEX IF NOT EXISTS symbol_monthly_tvl_timestamp ON symbol_monthly_tvl(timestamp);


CREATE TABLE IF NOT EXISTS symbol_cumulative_tvl (
      guid                VARCHAR PRIMARY KEY,
      symbol              VARCHAR NOT NULL,
      amount              UINT256 DEFAULT 0,
      latest_price        VARCHAR,
      trans_to_usd        UINT256 DEFAULT 0,
      timestamp           INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS symbol_cumulative_tvl_timestamp ON symbol_cumulative_tvl(timestamp);


CREATE TABLE IF NOT EXISTS protocol_daily_tvl (
    guid                VARCHAR PRIMARY KEY,
    protocol_id         VARCHAR NOT NULL,
    symbol              VARCHAR NOT NULL,
    amount              UINT256 DEFAULT 0,
    latest_price        VARCHAR,
    trans_to_usd        UINT256 DEFAULT 0,
    timestamp           INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS protocol_daily_tvl_timestamp ON protocol_daily_tvl(timestamp);


CREATE TABLE IF NOT EXISTS protocol_weekly_tvl (
     guid                VARCHAR PRIMARY KEY,
     protocol_id         VARCHAR NOT NULL,
     symbol              VARCHAR NOT NULL,
     amount              UINT256 DEFAULT 0,
     latest_price        VARCHAR,
     trans_to_usd        UINT256 DEFAULT 0,
     timestamp           INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS protocol_weekly_tvl_timestamp ON protocol_weekly_tvl(timestamp);


CREATE TABLE IF NOT EXISTS protocol_monthly_tvl (
    guid                VARCHAR PRIMARY KEY,
    protocol_id         VARCHAR NOT NULL,
    symbol              VARCHAR NOT NULL,
    amount              UINT256 DEFAULT 0,
    latest_price        VARCHAR,
    trans_to_usd        UINT256 DEFAULT 0,
    timestamp           INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS protocol_monthly_tvl_timestamp ON protocol_monthly_tvl(timestamp);


CREATE TABLE IF NOT EXISTS protocol_cumulative_tvl (
     guid                VARCHAR PRIMARY KEY,
     protocol_id         VARCHAR NOT NULL,
     symbol              VARCHAR NOT NULL,
     amount              UINT256 DEFAULT 0,
     latest_price        VARCHAR,
     trans_to_usd        UINT256 DEFAULT 0,
     timestamp           INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS protocol_cumulative_tvl_timestamp ON protocol_cumulative_tvl(timestamp);

CREATE TABLE IF NOT EXISTS token_lists (
     id             SERIAL PRIMARY KEY,
     chain_id       UINT256 NOT NULL,
     address        VARCHAR NOT NULL,
     name           VARCHAR,
     symbol         VARCHAR,
     decimals       UINT256 NOT NULL,
     timestamp      INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS token_lists_timestamp ON token_lists(timestamp);
