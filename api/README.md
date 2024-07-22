# API Docs

## Configuration

##### Responses

| http code | content-type               | response                       |
| --------- | -------------------------- | ------------------------------ |
| `2000`    | `text/plain;charset=UTF-8` | `Called successfully`          |
| `4000`    | `application/json`         | `Client input parameter error` |
| `5000`    | `text/html;charset=utf-8`  | `Server input parameter error` |

---

## API Endpoints

### Service Checks

<details>
 <summary><code>GET</code> <code><b>/healthz</b></code> <code>(Check that the service is in a healthy state)</code></summary>

##### Parameters

> None

##### Example cURL

> ```bash
>  curl -X GET -H "Content-Type: application/json" http://127.0.0.1:9090/healthz
> ```

</details>

<details>
 <summary><code>GET</code> <code><b>/metrics</b></code> <code>(Monitoring Metrics Interface)</code></summary>

##### Parameters

> None

##### Example cURL

> ```bash
>  curl -X GET -H "Content-Type: application/json" http://127.0.0.1:9090/metrics
> ```

</details>

---

### Query Methods

<details>
 <summary><code>GET</code> <code><b>/api/v1/deposits/</b></code> <code>(Query the list of deposit transactions by address and paging information)</code></summary>

##### Parameters

| Name       | Type    | Position   | Description    | Required                                                                                                                             |
| ---------- | ------- | ---------- | -------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `address`  | string  | Body Param | User's address | No. Input `0x00`, if the address is not null return the transaction associated with that address, otherwise return all transactions. |
| `page`     | Integer | Body Param | Page number    | No. Return to page 1 by default                                                                                                      |
| `pageSize` | Integer | Body Param | Page size      | No. Return to 20th data by default                                                                                                   |
| `order`    | string  | Body Param | Order          | Yes. `asc`: ascend order <br> `desc`：descend order                                                                                  |

##### Response

| Name                | Type    | Description                               |
| ------------------- | ------- | ----------------------------------------- |
| `l1TransactionHash` | string  | Layer1 deposit tx hash                    |
| `l2TransactionHash` | string  | Layer2 claim deposit tx hash              |
| `l1BlockNumber`     | uint256 | Layer1 block number                       |
| `status`            | uint8   | tx status: <br> `1`: pending; `2`:success |
| `l1TokenAddress`    | string  | Layer1 token address                      |
| `l2TokenAddress`    | string  | Layer2 token address                      |
| `fromAddress`       | string  | From address                              |
| `toAddress`         | string  | To address                                |
| `ETHAmount`         | uint256 | ETH amount                                |
| `ERC20Amount`       | uint256 | ERC20 amount                              |
| `blockTimestamp`    | uint256 | timestamp                                 |
| `queueIndex`        | uint256 | V1 deposit queue index                    |
| `l1TxOrigin`        | string  | L1 Tx Origin                              |
| `gasLimit`          | uint256 | Gas Limit                                 |

##### Example cURL

> ```bash
>  curl -X GET -H "Content-Type: application/json" \
>   -d '{"address":"0x00","page":1,"pageSize":20,"order":"asc"}' \
>   http://127.0.0.1:9090/api/v1/deposits/
> ```

</details>

<details>
 <summary><code>GET</code> <code><b>/api/v1/withdrawals/</b></code> <code>(Query the list of withdrawal transactions by address and paging information)</code></summary>

##### Parameters

| Name       | Type    | Position   | Description    | Required                                                                                                               |
| ---------- | ------- | ---------- | -------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `address`  | string  | Body Param | User's address | No. If the address is not null return the transaction associated with that address, otherwise return all transactions. |
| `page`     | Integer | Body Param | Page number    | No. Return to page 1 by default                                                                                        |
| `pageSize` | Integer | Body Param | Page size      | No. Return to 20th data by default                                                                                     |
| `order`    | string  | Body Param | Order          | Yes. `asc`: ascend order <br> `desc`：descend order                                                                    |

##### Response

| Name                | Type    | Description                                                                                                |
| ------------------- | ------- | ---------------------------------------------------------------------------------------------------------- |
| `l1TransactionHash` | string  | Layer1 withdraw tx hash                                                                                    |
| `l2TransactionHash` | string  | Layer2 claim withdraw tx hash                                                                              |
| `l1ProveTxHash`     | string  | Layer1 withdraw prove tx hash                                                                              |
| `l1BlockNumber`     | uint256 | Layer1 block number                                                                                        |
| `status`            | uint8   | tx status: <br> `1`: Waiting `2`:Ready to Prove `3`:In Challenge Period `4`:Ready to Finalized `5`:Relayed |
| `l1TokenAddress`    | string  | Layer1 token address                                                                                       |
| `l2TokenAddress`    | string  | Layer2 token address                                                                                       |
| `fromAddress`       | string  | From address                                                                                               |
| `toAddress`         | string  | To address                                                                                                 |
| `ETHAmount`         | uint256 | ETH amount                                                                                                 |
| `ERC20Amount`       | uint256 | ERC20 amount                                                                                               |
| `blockTimestamp`    | uint256 | timestamp                                                                                                  |
| `msgNonce`          | uint256 | Self-incrementing nonce in `CrossDomainMessage` contract                                                   |
| `timeLeft`          | uint256 | Left time of the challenge period                                                                          |

##### Example cURL

> ```bash
>  curl -X GET -H "Content-Type: application/json" \
>   -d '{"address":"0x00000000000000000000","page":1,"pageSize":20,"order":"desc"}' \
>   http://127.0.0.1:9090/api/v1/withdrawals/
> ```

</details>

<details>
 <summary><code>GET</code> <code><b>/api/v1/datastore/list/</b></code> <code>(Query the list of datastore by paging information)</code></summary>

##### Parameters

| Name       | Type    | Position   | Description | Required                                            |
| ---------- | ------- | ---------- | ----------- | --------------------------------------------------- |
| `page`     | Integer | Body Param | Page number | No. Return to page 1 by default                     |
| `pageSize` | Integer | Body Param | Page size   | No. Return to 20th data by default                  |
| `order`    | string  | Body Param | Order       | Yes. `asc`: ascend order <br> `desc`：descend order |

##### Response

| Name          | Type    | Description                                                                             |
| ------------- | ------- | --------------------------------------------------------------------------------------- |
| `DataStoreId` | Integer | MantleDA Datastore id                                                                   |
| `DataSize`    | string  | Batch size                                                                              |
| `Status`      | bool    | Data store status; <br> `True`: Data is valid in DA <br> `False`: Data is invalid in DA |
| `Age`         | Integer | Data store time                                                                         |
| `DaHash`      | string  | Transaction hash of data stored in MantleDA                                             |

##### Example cURL

> ```bash
>  curl -X GET -H "Content-Type: application/json" \
>   -d '{page":1,"pageSize":20,"order":"asc"}' \
>   http://127.0.0.1:9090/api/v1/datastore/list/
> ```

</details>

<details>
 <summary><code>GET</code> <code><b>/api/v1/datastore/id/{id}</b></code> <code>(Query the datastore by index id information)</code></summary>

##### Parameters

| Name | Type    | Position    | Description        | Required |
| ---- | ------- | ----------- | ------------------ | -------- |
| `id` | Integer | Query Param | Datastore Index id | Yes.     |

##### Response

| Name                    | Type     | Description                                                                             |
| ----------------------- | -------- | --------------------------------------------------------------------------------------- |
| `dataStoreId`           | Integer  | The datastore ID on MantleDA                                                            |
| `dataStoreNumber`       | Integer  | The datastore number on MantleDA                                                        |
| `durationDataStoreId`   | Integer  | The duration ID on MantleDA                                                             |
| `dataSize`              | integer  | The data size on MantleDA                                                               |
| `index`                 | Integer  | The data index number                                                                   |
| `dataCommitment`        | string   | The data commitment information                                                         |
| `msgHash`               | string   | The Msg hash of the data store                                                          |
| `initTime`              | Integer  | The init time of the data store                                                         |
| `expireTime`            | Integer  | The expire time of the data store                                                       |
| `duration`              | Integer  | Configured datastore expiration time                                                    |
| `status`                | bool     | Data store status; <br> `True`: Data is valid in DA <br> `False`: Data is invalid in DA |
| `numSys`                | uint8    | The number of the system nodes                                                          |
| `numPar`                | uint8    | The number of the partner nodes                                                         |
| `degree`                | string   | The degree of the KZG low-degree proof                                                  |
| `storePeriodLength`     | Integer  | The valid period length of the data store                                               |
| `fee`                   | Integer  | The data store fee                                                                      |
| `confirmer`             | string   | The confirmed address of the data store                                                 |
| `header`                | string   | The data store header                                                                   |
| `initTxHash`            | string   | The initial transaction Hash of the data                                                |
| `initGasUsed`           | string   | The initial transaction gasUsed of the data                                             |
| `initBlockNumber`       | Integer  | The initial transaction block number of the data                                        |
| `ethSigned`             | string   | The data commitment of the Ethereum signature.                                          |
| `mantleSigned`          | string   | The data commitment of the Mantle signature.                                            |
| `nonSignerPubKeyHashes` | []string | The list of the non-Signer public key hashes                                            |
| `signatoryRecord`       | string   | The signature record of the data commitment                                             |
| `confirmTxHash`         | string   | The confirmed transaction Hash of the data                                              |
| `confirmGasUsed`        | Integer  | The confirmed transaction gasUsed of the data                                           |

##### Example cURL

> ```bash
>  curl -X GET -H "Content-Type: application/json" http://127.0.0.1:9090/api/v1/datastore/id/100
> ```

</details>

<details>
 <summary><code>GET</code> <code><b>/api/v1/datastore/transaction/id/{id}</b></code> <code>(Query the transaction information stored in the corresponding datastore based on the id.)</code></summary>

##### Parameters

| Name | Type    | Position    | Description        | Required |
| ---- | ------- | ----------- | ------------------ | -------- |
| `id` | Integer | Query Param | Datastore Index id | Yes.     |

##### Response

| Name          | Type    | Description                                       |
| ------------- | ------- | ------------------------------------------------- |
| `storeId`     | Integer | The id of datastore                               |
| `index`       | Integer | The transaction index                             |
| `blockNumber` | Integer | The block number                                  |
| `txHash`      | string  | The transaction hash                              |
| `blockData`   | string  | The original data submitted to the DA in MantleV2 |

##### Example cURL

> ```bash
>  curl -X GET -H "Content-Type: application/json" http://127.0.0.1:9090/api/v1/datastore/transaction/id/10
> ```

</details>

<details>
 <summary><code>GET</code> <code><b>/api/v1/stateroot/list</b></code> <code>(Query the list of state root by paging information)</code></summary>

##### Parameters

| Name       | Type    | Position   | Description | Required                                            |
| ---------- | ------- | ---------- | ----------- | --------------------------------------------------- |
| `page`     | Integer | Body Param | Page number | No. Return to page 1 by default                     |
| `pageSize` | Integer | Body Param | Page size   | No. Return to 20th data by default                  |
| `order`    | string  | Body Param | Order       | Yes. `asc`: ascend order <br> `desc`：descend order |

##### Response

| Name                | Type    | Description                                                           |
| ------------------- | ------- | --------------------------------------------------------------------- |
| `blockHash`         | string  | Layer1 block hash                                                     |
| `transactionHash`   | string  | Layer1 transaction Hash                                               |
| `l1BlockNumber`     | uint256 | Layer1 block number                                                   |
| `l2BlockNumber`     | uint256 | Layer2 block number                                                   |
| `outputIndex`       | uint256 | The state root index                                                  |
| `prevTotalElements` | uint256 | The last count of elements                                            |
| `status`            | string  | The Layer1 status                                                     |
| `transactionHash`   | string  | The Layer1 transaction hash                                           |
| `outputRoot`        | string  | The state root of the Rollup data                                     |
| `canonical`         | bool    | The data derivation status: <br> `true`: Normal <br> `false`: Unmoral |
| `batchSize`         | uint256 | The batch size of the data                                            |
| `timestamp`         | uint256 | Timestamp                                                             |

##### Example cURL

> ```bash
>  curl -X GET -H "Content-Type: application/json" \
>   -d '{"page":1,"pageSize":20,"order":"asc"}' \
>   http://127.0.0.1:9090/api/v1/stateroot/list/
> ```

</details>

<details>
 <summary><code>GET</code> <code><b>/api/v1/stateroot/index/{index}</b></code> <code>(Query the datastore by index id information)</code></summary>

##### Parameters

| Name    | Type    | Position    | Description | Required |
| ------- | ------- | ----------- | ----------- | -------- |
| `index` | Integer | Query Param | Index id    | Yes.     |

##### Response

| Name                | Type    | Description                                                           |
| ------------------- | ------- | --------------------------------------------------------------------- |
| `blockHash`         | string  | Layer1 block hash                                                     |
| `transactionHash`   | string  | Layer1 transaction Hash                                               |
| `l1BlockNumber`     | uint256 | Layer1 block number                                                   |
| `l2BlockNumber`     | uint256 | Layer2 block number                                                   |
| `outputIndex`       | uint256 | The state root index                                                  |
| `prevTotalElements` | uint256 | The last count of elements                                            |
| `status`            | string  | The Layer1 status                                                     |
| `transactionHash`   | string  | The Layer1 transaction hash                                           |
| `outputRoot`        | string  | The state root of the Rollup data                                     |
| `canonical`         | bool    | The data derivation status: <br> `true`: Normal <br> `false`: Unmoral |
| `batchSize`         | uint256 | The batch size of the data                                            |
| `timestamp`         | uint256 | Timestamp                                                             |

##### Example cURL

> ```bash
>  curl -X GET -H "Content-Type: application/json" http://127.0.0.1:9090/api/v1/datastore/id/100
> ```

</details>
