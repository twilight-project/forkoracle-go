package main

import (
	"os"
	"log"
)

func main() {

	accountName := os.Args[1]
	orchestrator(accountName)

	file, err := openLogFile("Debug.log")
    	if err != nil {
        	log.Fatal(err)
    	}
    	log.SetOutput(file)
    	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

    	log.Println("log file created")
}


func openLogFile(path string) (*os.File, error) {
    logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
    if err != nil {
        return nil, err
    }
    return logFile, nil
}
