CREATE TABLE IF NOT EXISTS relay_message (
   guid                          VARCHAR PRIMARY KEY,
   block_number                  UINT256 NOT NULL,
   relay_transaction_hash        VARCHAR NOT NULL,
   deposit_hash                  VARCHAR NOT NULL,
   message_hash                  VARCHAR,
   l1_token_address              VARCHAR,
   l2_token_address              VARCHAR,
   eth_amount                    UINT256,
   erc20_amount                  UINT256,
   related                       BOOLEAN DEFAULT FALSE,
   timestamp                     INTEGER NOT NULL CHECK (timestamp > 0)
);
CREATE INDEX IF NOT EXISTS relay_message_message_hash ON relay_message(message_hash);
CREATE INDEX IF NOT EXISTS relay_message_timestamp ON relay_message(timestamp);
