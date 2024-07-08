
# simple GoRedisClone
This is a simple Redis clone written in go that supports keys and lists with the client library, it uses already written [go-library](https://github.com/tidwall/resp) for resp and [testify](https://github.com/stretchr/testify) for testing.

## Features

- concurrent writes and reads
- logs
- password support for client and server
- data recovery for persistent data 
- context support for the client
- databases support (40)

## How to run

To run this project you can do multiple things 


#### Run in the Main repository
```bash
  make run
```

#### Run in the Docker container
#### Important!
Make sure you don't have images with the name rediscl and containers with the name myredis
#### To stop the container and delete the image  
```bash
  make docker_rm 
```
## Why I made it 
I'm relatively new in programming and go, so I decided to write Redis which I used in my previous projects, but simple and small. I found it really helpful and interesting, would like to see any feedback on it. Thanks! 