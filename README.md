# Blockchain War Game

This is an example server project of zecrey `Blockchain War`.

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
   --data '{"query":"query MyQuery {collection(where: {account: {account_name: {_eq: \"gavinplaygameserver2.zec\"}}, l2_collection_id: {_eq: \"0\"}}) {id}}","variables":{}}'

#{"data":{"collection":[{"id":5}]}}
```

We use docker-compose to start the service. Please refer to [here](https://docs.docker.com/compose/install/) for
docker-compose installation

Finally, run the development server:

```bash
  cd BlockChainWar/
  docker-compose -f docker-compose.yaml up -d
```
