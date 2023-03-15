# Zecrey Block war

This is an example server project of Zecrey Block war.

## Getting Started

If you do not have docker installed, [install docker](https://dockerdocs.cn/desktop/#download-and-install).

Clone the repo and create the `config.yaml` file

```bash
  cd ZecreyBlockWar/game/api/server/etc/ && cp config.yaml.example config.yaml
```

Modify the `config.yaml` file to configure your information, following is an example:

```json

{
  "database": {
    "host": "localhost",
    "port": 5433,
    "user": "root",
    "password": "public",
    "database": "zecreyBlockWar"
  },
  "fps": 30,
  "game_round_interval": 0,
  "frontend_type": "zecrey_warrior",
  "item_frame_chance": 500,
  "game_duration": 60,
  "account_name": "amber1",
  "seed": "<private_key_from_metamask>",
  "nft_prefix": "goodwei",
  "collection_id": "<ID of the collection you created>"
}


```

Each account has a collection created by default. You can query through this example curl

You can replace `amber1.zec` in example with your name for query.

Example:

```bash
   curl --location 'https://hasura.zecrey.com/v1/graphql' \
   --header 'Content-Type: application/json' \
   --data '{"query":"query MyQuery {\n  collection(where: {account: {account_name: {_eq: \"gavinplaygameserver2.zec\"}}, l2_collection_id: {_eq: \"0\"}}) {\n    id\n  }\n}","variables":{}}'
```

Example result:

```bash
 #{"data":{"collection":[{"id":5}]}}
```

Then,run the development server:

```bash
  docker-compose -f docker-compose.yaml up -d
```

If your docker download image is stuck,
please [configure Docker image source](https://mirrors.ustc.edu.cn/help/dockerhub.html#linux)
