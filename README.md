# Block Chain War Game

This is an example server project of zecrey 'Block Chain War'.

## Getting Started

If you do not have docker installed, [install docker](https://dockerdocs.cn/desktop/#download-and-install).

Clone the repo and create the `config.json` file

```bash
  cd BlockChainWar/config/ && cp config.json.example config.json
```

Modify the `config.yaml` file to configure your information, following is an example:

```json
    {
      "database": {                     
        "host": "postgres",            //If you do not use docker-compose to start,please modify the host specified for you here
        "port": 5432,
        "user": "root",
        "password": "public",
        "database": "block_chain_war"
      },                              
      "fps": 30,
      "game_round_interval": 0,
      "frontend_type": "block_chain_war",
      "item_frame_chance": 500,
      "game_duration": 60,              //Duration of a game (s)
      "seed": "<private_key_from_metamask>",
      "nft_prefix": "companyName",
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

We use docker-compose to start the service. Please refer to [here](https://docs.docker.com/compose/install/) for
docker-compose installation

docker-compose.yaml analysis:

```bash
version: '3.9'

services:
  postgres:
    image: postgres:13.4-alpine3.14
    hostname: block-chain-war-postgres
    container_name: block-chain-war-postgres
    ports:
      - "5433:5432"
    environment:
      - POSTGRES_PASSWORD=public
      - POSTGRES_DB=block_chain_war
      - POSTGRES_USER=root
    restart: unless-stopped

  zecrey_war:
    image: zecrey/zecrey-chain-war:0.0.6
    hostname: block-chain-war
    container_name: block-chain-war
    ports:
      - "3250:3250"
      - "3251:3251"
    volumes:
      - ./config/config.json:/block-chain-war/config/config.json    
      //We want to use our own configuration, so please check whether the configuration in database in 'config. json' 
        is consistent with the configuration in 'postgres' under' services: '
    depends_on:
      - postgres
    command: [ "./wait-for-it.sh", "postgres:5432", "--", "./main" ]
```

Then,run the development server:

```bash
  cd BlockChainWar/
  docker-compose -f docker-compose.yaml up -d
```

If your docker download image is stuck,
please [configure Docker image source](https://mirrors.ustc.edu.cn/help/dockerhub.html#linux)
