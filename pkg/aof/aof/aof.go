package AOF

import (
	"bufio"
	"fmt"
	"io"
	"os"
	RespWriter "saidis/pkg/writers/respWriter"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Aof struct {
	file *os.File      // file to read
	rd   *bufio.Reader // reader
	mu   sync.Mutex    // mutex to read concurrently and prevent the read opertaion multiple time by the threads
}

func NewAof(path string) (*Aof, error) {
	// open file or created if does not exists

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	aof := &Aof{
		file: f,
		rd:   bufio.NewReader(f),
	}

	// Start a goroutine to sync AOF to disk every 1 second
	go func() {
		for {
			aof.mu.Lock()

			aof.file.Sync()

			aof.mu.Unlock()

			time.Sleep(time.Second)
		}
	}()

	return aof, nil
}

func (aof *Aof) Close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	return aof.file.Close()
}

// write the command to the file in the same RESP format that we receive. This way, when we read the file later, we can parse these RESP lines and write them back to memory.
func (aof *Aof) Write(value RespWriter.Value) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	_, err := aof.file.Write(value.Marshal())
	if err != nil {
		return err
	}

	return nil
}

// read the commands from the file then execute them
func (aof *Aof) Read(callback func(value RespWriter.Value)) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	// Reset the file pointer to the beginning of the file
	_, err := aof.file.Seek(0, 0)
	if err != nil {
		return err
	}

	// Read the data from the file
	reader := bufio.NewReader(aof.file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break // End of file reached
			}
			return err // Return any other error
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "*") {
			// Number of arguments
			numArgs, err := strconv.Atoi(line[1:])
			if err != nil {
				return err
			}

			// Read each argument
			args := make([]RespWriter.Value, numArgs)
			for i := 0; i < numArgs; i++ {
				// Read the length of the bulk string
				bulkLengthLine, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				bulkLengthLine = strings.TrimSpace(bulkLengthLine)

				if !strings.HasPrefix(bulkLengthLine, "$") {
					return fmt.Errorf("expected bulk string length, got: %s", bulkLengthLine)
				}

				bulkLength, err := strconv.Atoi(bulkLengthLine[1:])
				if err != nil {
					return err
				}

				// Read the actual bulk string
				bulkData := make([]byte, bulkLength)
				_, err = io.ReadFull(reader, bulkData)
				if err != nil {
					return err
				}

				// Read the newline after the bulk data
				_, err = reader.ReadString('\n')
				if err != nil {
					return err
				}

				// Create a Value for the argument
				args[i] = RespWriter.Value{Bulk: string(bulkData)}
			}

			// Call the callback function with the array of values
			callback(RespWriter.Value{Array: args})
		}
	}

	return nil
}
