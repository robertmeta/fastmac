package main

import "log"
import "os"
import "bufio"
import "sync"

const (
	program = "fastmac"
	version = "0.1"
)

var (
	debug bool
	fetchDebug sync.Once
	logFilename string
	fetchLogfilename sync.Once
)

func main() {
	if logToFile() != "" {
		f, err := os.OpenFile(logToFile(), os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		log.SetOutput(f)
	}

	if debugMode() {
		log.Println("Starting fastmac server")
	}

	err := NsSpeechInit()
	if err != nil {
		log.Fatal("Failed to start speech server")
	}
	defer NsSpeechFree()

	
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		t := scanner.Text()
		if debugMode() {
			log.Println("parsing: ", t)
		}
		
		err = processLine(t)
		if err != nil {
			if debugMode() {
				log.Println("failed to parse line", err)
			}
		}
	}

	if scanner.Err() != nil {
		log.Fatal("Exiting "+program+" server with scanner error: ", scanner.Err())
	}
	
	if debugMode() {
		log.Println("Exiting "+program+" server")
	}
}


func debugMode() bool {
	fetchDebug.Do(func() {
		ldebug := os.Getenv("DEBUG")
		if ldebug != "" {
			debug = true
		}
		// else debug is already false
	})
	return debug
	
}

func logToFile() string {
	fetchLogfilename.Do(func() {
		logFilename = os.Getenv("LOG")
	})
	return logFilename
}


