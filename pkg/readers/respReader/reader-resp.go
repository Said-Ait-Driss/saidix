package RespReaders

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	RespWriter "saidis/pkg/writers/respWriter"
	"strconv"
)

/*
*  RESP string : $5\r\nAhmed\r\n
* RESP aray : *2\r\n$5\r\nhello\r\n$5\r\nworld\r\n
* samplified RESP aray :
* *2
* $5
* hello
* $5
* world

 */

/*
 $ --> the type of string
 * --> the type of array
 \r\n --> the end of line

*/

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

// typ :  is used to determine the data type carried by the value.
// str : holds the value of the string received from the simple strings.
// num : holds the value of the integer received from the integers.
// bulk : is used to store the string received from the bulk strings.
// array : holds all the values received from the arrays.

/*
	  example of Value object
		Value{
			typ: "array",
			array: []Value{
				Value{typ: "bulk", bulk: "SET"},
				Value{typ: "bulk", bulk: "name"},
				Value{typ: "bulk", bulk: "Ahmed"},
			},
		}
*/

func UnmarshalValue(data []byte) (RespWriter.Value, error) {
	var value RespWriter.Value
	err := json.Unmarshal(data, &value)
	fmt.Println("data : ", value)

	if err != nil {
		return RespWriter.Value{}, err
	}
	return value, nil
}

// The reader struct
type Resp struct {
	Reader *bufio.Reader
}

// init the reader
func NewResp(rd io.Reader) *Resp {
	return &Resp{Reader: bufio.NewReader(rd)}
}

/*
readline :

	read byte by byte then
	return number of bytes => n
	return the actual string without the two latest characters ( /r/n ) => line[:len(line)-2]
*/
func (r *Resp) ReadLine() (line []byte, n int, err error) {
	for {
		b, err := r.Reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	return line[:len(line)-2], n, nil
}

/*
readInteger :
	read the integer from the reader
	return the integer => i64
	return the number of bytes read => n
*/

func (r *Resp) ReadInteger() (x int, n int, err error) {
	line, n, err := r.ReadLine()
	if err != nil {
		return 0, 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i64), n, nil
}

/*
	read the lenght of array
	read the value for each line of the array
	append each value to array then return the array
	final result returned : {
								"array" : [value1, value2, value3, ...]
								"typ" : "array"
							}
*/

func (r *Resp) ReadArray() (RespWriter.Value, error) {
	v := RespWriter.Value{}
	v.Typ = "Array"

	// read length of array
	len, _, err := r.ReadInteger()
	if err != nil {
		return v, err
	}

	// foreach line, parse and read the value
	v.Array = make([]RespWriter.Value, 0)
	for i := 0; i < len; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}

		// append parsed value to array
		v.Array = append(v.Array, val)
	}

	return v, nil
}

/*
	read the lenght of string
	read the string
	use readline to read the entire string and return the it without the two latest characters ( /r/n )
	return the string
	the final result returned : {
									"string" : "string_value"
									"typ" : "string"
								}
*/

func (r *Resp) ReadBulk() (RespWriter.Value, error) {
	v := RespWriter.Value{}

	v.Typ = "Bulk"

	// read length of string
	len, _, err := r.ReadInteger()
	if err != nil {
		return v, err
	}

	bulk := make([]byte, len)

	r.Reader.Read(bulk)

	v.Bulk = string(bulk)

	// Read the trailing CRLF
	r.ReadLine()

	return v, nil
}

/*
	read the first byte to determine the type of arg
	proccess the arg based on the type
*/

func (r *Resp) Read() (RespWriter.Value, error) {
	_type, err := r.Reader.ReadByte()

	if err != nil {
		return RespWriter.Value{}, err
	}

	switch _type {
	case ARRAY:
		return r.ReadArray()
	case BULK:
		return r.ReadBulk()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return RespWriter.Value{}, nil
	}
}
