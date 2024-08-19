package main

import (
	"fmt"
	"io"
	"net"
	"os"
	AOF "saidis/pkg/aof/aof"
	handlers "saidis/pkg/handlers/aofHandler"
	RespReaders "saidis/pkg/readers/respReader"
	RespWriter "saidis/pkg/writers/respWriter"
	"strconv"
	"strings"
)

func main() {
	port := 6379
	fmt.Println("saidis server listen on port :", port)
	server, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println(err)
		return
	}

	aof, err := AOF.NewAof("database.aof")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer aof.Close()

	aof.Read(func(value RespWriter.Value) {
		command := strings.ToUpper(value.Array[0].Bulk)

		args := value.Array[1:]

		handler, ok := handlers.Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			return
		}

		handler(args)
	})

	con, err := server.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	defer con.Close()
	// creating infiniti loop to receive requests from client and respond to them
	for {

		res := RespReaders.NewResp(con)
		value, err := res.Read()

		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("error reading from client: ", err.Error())
			os.Exit(1)
		}

		if value.Typ != "Array" {
			fmt.Println("Invalid request, expected Array")
			continue
		}

		if len(value.Array) == 0 {
			fmt.Println("Invalid request, expected Array length > 0")
			continue
		}
		command := strings.ToUpper(value.Array[0].Bulk)
		args := value.Array[1:]

		writer := RespWriter.NewWriter(con)

		handler, ok := handlers.Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			writer.Write(RespWriter.Value{Typ: "string", Str: ""})
			continue
		}
		if command == "SET" || command == "HSET" {
			aof.Write(value)
		}

		result := handler(args)
		writer.Write(result)

	}
}
