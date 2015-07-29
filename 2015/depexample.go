package main

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
)

func main() {
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	conn.Do("SET", "message", "Hello World")
	world, err := redis.String(conn.Do("GET", "message"))
	if err != nil {
		fmt.Println("key not found")
	}

	fmt.Println(world)
}
