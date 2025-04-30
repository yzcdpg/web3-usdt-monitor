# Web3 USDT Monitor

## Features
- [x] 监控波长`USDT`转账
- [ ] 动态增加用户

## Config
在`last_block.json`中配置开始区块，如果不配置，则从0开始。 结束时，会将最后扫描的区块写入该文件。

## Build & Run
```
go run main.go
```

## Tron Transaction
```json

{
  "blockID": "0000000004466983e78b77fc3879e7ddbe1de5215ac7d55cd1ccbb6f59893230",
  "block_header": {
    "raw_data": {
      "number": 71723395,
      "txTrieRoot": "e48d0857b1b2ef05c21acf940ce3d806fe9f7b2f1aa44253ea00cfda0a4d3044",
      "witness_address": "418440ffd578f7a5abf3537b5f46a6980d382db581",
      "parentHash": "00000000044669824a0d4c8beabb227b089144f7104fbd226b1fee63727dd279",
      "version": 31,
      "timestamp": 1745825274000
    },
    "witness_signature": "d31efabe853c777cedac752f31d658bac764238ecc264d7b5ab92f5bd70fa19012fd09fa10cb7a78c69dba496c229ec96405edb0abdf727bad2b244d792f29ff01"
  },
  "transactions": [
    {
      "ret": [
        {
          "contractRet": "SUCCESS"
        }
      ],
      "signature": [
        "5d22c9f790932b192f6603abd7728d3f3431eef0869d422c85c3ce527bb583714e67c8077662bed072148adf27d5142cfdfdf3cb7f79c1325aa72a28f5b2b72400"
      ],
      "txID": "369e1efdeedfc774f967aff173ab22a1ce58471f08e624a190aac4d148f1d943",
      "raw_data": {
        "contract": [
          {
            "parameter": {
              "value": {
                "amount": 1,
                "owner_address": "41f96c4705e780785f6e46e84da9744105f0c2142a",
                "to_address": "4155181e20f78902ed2ff8fc7a75814762887d497b"
              },
              "type_url": "type.googleapis.com/protocol.TransferContract"
            },
            "type": "TransferContract"
          }
        ],
        "ref_block_bytes": "696f",
        "ref_block_hash": "1ef122b9b047568a",
        "expiration": 1745825328000,
        "timestamp": 1745825270741
      },
      "raw_data_hex": "0a02696f22081ef122b9b047568a4080d7b1dae7325a65080112610a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412300a1541f96c4705e780785f6e46e84da9744105f0c2142a12154155181e20f78902ed2ff8fc7a75814762887d497b180170d597aedae732"
    },
    {
      "ret": [
        {
          "contractRet": "SUCCESS"
        }
      ],
      "signature": [
        "ef42241c10af2107f175e166401ffdb25c4cdd96f3ba02742194e281cde3ffac1d8c85a06e252adb6cd0072f722d3ec50571ea8c835c6312fb272aaa75f9d58e1b"
      ],
      "txID": "ecd50059428170eee5e4c9862ad56f1d582d26540cecc071899eef7d07e19d38",
      "raw_data": {
        "contract": [
          {
            "parameter": {
              "value": {
                "data": "a9059cbb0000000000000000000000003b9f381f7e5017f3966277d00a0ca5adfc5634920000000000000000000000000000000000000000000000000000000000989680",
                "owner_address": "415a274493e64a26e6bbf38e86d42192045f4e6881",
                "contract_address": "41a614f803b6fd780986a42c78ec9c7f77e6ded13c"
              },
              "type_url": "type.googleapis.com/protocol.TriggerSmartContract"
            },
            "type": "TriggerSmartContract"
          }
        ],
        "ref_block_bytes": "6806",
        "ref_block_hash": "7b19b694990c0f43",
        "expiration": 1745911369682,
        "fee_limit": 30000000,
        "timestamp": 1745825269682
      },
      "raw_data_hex": "0a02680622087b19b694990c0f4340d29fb583e8325aae01081f12a9010a31747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e54726967676572536d617274436f6e747261637412740a15415a274493e64a26e6bbf38e86d42192045f4e6881121541a614f803b6fd780986a42c78ec9c7f77e6ded13c2244a9059cbb0000000000000000000000003b9f381f7e5017f3966277d00a0ca5adfc563492000000000000000000000000000000000000000000000000000000000098968070b28faedae73290018087a70e"
    }
  ]
}
```


