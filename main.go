package main

import (
	"errors"
	"flag"

	"github.com/Som-Kesharwani/shared-service/logger"
)

func main() {
	//Parse log level Info from command line
	logLevel := flag.Int("loglevel", 1, "an integer value (0-4)")
	flag.Parse()
	//calling the SetLogLevel with the comand-line argument
	logger.SetLogLevel(logger.Level(*logLevel), "Mylog.text")
	flag.Parse()
	//Calling the SetLogLevel with the command-line argument
	logger.Trace.Println("Main Started")
	loop()
	err := errors.New("text string")
	logger.Error.Println(err.Error())
	logger.Trace.Println("Main Completed")

}

func loop() {
	logger.Trace.Println("Loop startes")

	for i := 0; i < 10; i++ {
		logger.Info.Printf("Counter value is : %d", i)
	}
	logger.Warning.Printf("The counter variable is not being userd")
	logger.Trace.Println("Loop Completed!!")

}
