# example
There are 2 simple applications doing similar thing. One is written in NodeJS and another in Golang.

Both require working Redis server to keep session & timestamp.


## Installation ##

## hello-1 ##

```
npm install express
npm install redis

export port=redis_port
export host=redis_host
export password=redis_password

node hello1/hello-1.js
```

## hello-2 ##

```
cd hello2

go mod download

export REDIS_HOST="localhost:6379"
export REDIS_PASSWORD="password"
export REDIS_DB="0"

go run main.go
```
