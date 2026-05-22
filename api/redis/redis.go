package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

var Client = goredis.NewClient(&goredis.Options{
	Addr: "localhost:6379",
})
