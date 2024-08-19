package handlers

import (
	RespWriter "saidis/pkg/writers/respWriter"
	"sync"
)

/*
Redis commands are case-insensitive.
*/
var Handlers = map[string]func([]RespWriter.Value) RespWriter.Value{
	"PING": ping,
	"SET":  set,
	"GET":  get,
	"HSET": hset,
	"HGET": hget,
	// "HGETALL": hgetall,
}

var SETs = map[string]string{}
var SETsMu = sync.RWMutex{}

// response of ping command
func ping(args []RespWriter.Value) RespWriter.Value {
	if len(args) == 0 {
		return RespWriter.Value{Typ: "string", Str: "PONG"}
	}

	return RespWriter.Value{Typ: "string", Str: args[0].Bulk}
}

// response of set command to set Hash map (map[string]string)
// use RWMutex for handling requests concurrently and to ensure that the SETs map is not modified by multiple threads at the same time.
func set(args []RespWriter.Value) RespWriter.Value {
	if len(args) != 2 {
		return RespWriter.Value{Typ: "error", Str: "ERR wrong number of arguments for 'set' command"}
	}

	key := args[0].Bulk
	value := args[1].Bulk

	SETsMu.Lock()
	SETs[key] = value
	SETsMu.Unlock()

	return RespWriter.Value{Typ: "string", Str: "OK"}
}

// response of get command
func get(args []RespWriter.Value) RespWriter.Value {
	if len(args) != 1 {
		return RespWriter.Value{Typ: "error", Str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].Bulk

	SETsMu.RLock()
	value, ok := SETs[key]
	SETsMu.RUnlock()

	if !ok {
		return RespWriter.Value{Typ: "null"}
	}

	return RespWriter.Value{Typ: "bulk", Bulk: value}
}

// The HSET & HGET commands
/*
	a Hash Map within a Hash Map:
		map[string]map[string]string
	example :
		- HSET users u1 Ahmed
		- HGET users u1

*/

var HSETs = map[string]map[string]string{}
var HSETsMu = sync.RWMutex{}

func hset(args []RespWriter.Value) RespWriter.Value {
	if len(args) != 3 {
		return RespWriter.Value{Typ: "error", Str: "ERR wrong number of arguments for 'hset' command"}
	}

	hash := args[0].Bulk
	key := args[1].Bulk
	value := args[2].Bulk

	HSETsMu.Lock()
	if _, ok := HSETs[hash]; !ok {
		HSETs[hash] = map[string]string{}
	}
	HSETs[hash][key] = value
	HSETsMu.Unlock()

	return RespWriter.Value{Typ: "string", Str: "OK"}
}

func hget(args []RespWriter.Value) RespWriter.Value {
	if len(args) != 2 {
		return RespWriter.Value{Typ: "error", Str: "ERR wrong number of arguments for 'hget' command"}
	}

	hash := args[0].Bulk
	key := args[1].Bulk

	HSETsMu.RLock()
	value, ok := HSETs[hash][key]
	HSETsMu.RUnlock()

	if !ok {
		return RespWriter.Value{Typ: "null"}
	}

	return RespWriter.Value{Typ: "bulk", Bulk: value}
}

// func hgetall(args []Value) Value {
// 	if len(args) != 2 {
// 		return Value{typ: "error", str: "ERR wrong number of arguments for 'hgetall' command "}
// 	}

// }
