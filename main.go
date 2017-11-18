package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

func checkErrorAndExit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

type ApacheErrorLog struct {
	dateTime time.Time
	clientIP net.IP
	message  string
}

func main() {
	var apacheErrorLogDateRegex = regexp.MustCompile(`\[([^\]]+)\]`)
	var apacheErrorLogClientRegex = regexp.MustCompile(`(?:\[client ([^\]]+)\])`)
	var apacheErrorLogRegex = regexp.MustCompile(`^\[([^\]]+)\] \[([^\]]+)\] (?:\[client ([^\]]+)\])? *(.*)$`)

	file, err := os.OpenFile("/var/log/apache2/error.log", os.O_RDONLY, 0666)

	checkErrorAndExit(err)

	defer file.Close()

	// This would be an awesome place to create an ApacheLogScanner. It would ignore
	// lines which don't represent legit apache log lines.
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 {
			continue
		}

		if !apacheErrorLogRegex.MatchString(line) {
			fmt.Println("Unrecognized log format", line)
		} else {
			log := ApacheErrorLog{}

			lineBytes := []byte(line)

			// Parse out the dates.
			apacheLogDateBytes := apacheErrorLogDateRegex.Find(lineBytes)

			if apacheLogDateBytes != nil {
				dateTime, err := time.Parse(time.ANSIC, strings.Trim(string(apacheLogDateBytes), "[]"))

				if err != nil {
					fmt.Println("Failed to parse log datetime.")
				} else {
					log.dateTime = dateTime
				}
			}

			// Parse out the client IP, if any.
			apacheLogClientBytes := apacheErrorLogClientRegex.Find(lineBytes)

			if apacheLogClientBytes != nil {
				clientString := string(apacheLogClientBytes)

				// TODO: There must be a way to trip the IP out via the regex.
				clientString = strings.TrimLeft(clientString, "[client ")
				clientString = strings.TrimRight(clientString, "]")

				log.clientIP = net.ParseIP(clientString)
			}

			// Parse out the error message.
			// TODO: Is it safe to assume that everything after the last ']' is the message?
			if cut := strings.LastIndexByte(line, ']'); cut != -1 {
				log.message = strings.TrimSpace(line[cut+1:])
			}
		}
	}

	// We've stopped scanning, that's either due to an error or EOF.
	if err = scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading log input:", err)
	}
}
