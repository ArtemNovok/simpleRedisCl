package main

import (
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:6666")
	if err != nil {
		panic(err)
	}

	// raw := "*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$3\r\nbar\r\n"
	conn.Write([]byte("skdfhksdfhk"))
}
