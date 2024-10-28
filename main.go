package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const layout string = "2006-01-02 15:04:05.999"
var waitGroup = sync.WaitGroup{}

type LogMessage struct {
	timestamp string
	severity string
	module string
	function string
	lineNumber int64
	message string
}

type LogAnalysis struct {
	numEntries int
	logSeverityFrequency LogSeverityFrequency
	topFiveLogMessages []string
	topFiveLogMessageFrequencies []int64
	startTime time.Time
	endTime time.Time
}

type LogSeverityFrequency struct {
	debug int64
	info int64
	warning int64
	error int64
}

func parseLogMessage(logRow string) (LogMessage, error) {
	var logMessage LogMessage
	leftParts := strings.Split(logRow, "|")
	if len(leftParts) != 3 {
		return logMessage, errors.New("Empty Message")
	}
	logMessage.timestamp = strings.TrimSpace(leftParts[0])
	logMessage.severity = strings.TrimSpace(leftParts[1])
	if logMessage.severity == "" {
		return logMessage, errors.New("Malformed message")
	}
	rightParts := strings.Split(leftParts[2], ":")
	if len(rightParts) < 3 {
		return logMessage, errors.New("Malformed message")
	}
	logMessage.module = strings.TrimSpace(rightParts[0])
	logMessage.function = strings.TrimSpace(rightParts[1])
	messageRaw := strings.Split(rightParts[2], "-")
	if len(messageRaw) < 2 {
		return logMessage, errors.New("Malformed message")
	}
	lineNumRaw := strings.Split(rightParts[2], "-")[0]
	message := strings.Split(rightParts[2], "-")[1]
	lineNum, err := strconv.ParseInt(strings.TrimSpace(lineNumRaw), 0, 16)
	logMessage.lineNumber = lineNum
	logMessage.message = strings.TrimSpace(message)
	if err != nil {
		return logMessage, err
	}
	return logMessage, nil
}

func parseLogFile(logPath string) (logMessages []LogMessage) {
	//waitGroup := sync.WaitGroup{}
	data, err := os.ReadFile(logPath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	logRows := strings.Split(string(data), "\n")
	for _, logRow := range logRows {
		logMessage, err := parseLogMessage(logRow)
		if err == nil {
			logMessages = append(logMessages, logMessage)
		}
	}
	return
}

func getNumEntries(logMessages []LogMessage) (numLogMessages int) {
	numLogMessages = len(logMessages)
	return
}

func getLogSeverityFrequency(logMessages []LogMessage) (logSeverityFrequency LogSeverityFrequency) {
	for _, logMessage := range logMessages {
		switch {
			case logMessage.severity == "DEBUG":
				logSeverityFrequency.debug += 1
			case logMessage.severity == "INFO":
				logSeverityFrequency.info += 1
			case logMessage.severity == "WARNING":
				logSeverityFrequency.warning += 1
			case logMessage.severity == "ERROR":
				logSeverityFrequency.error += 1
			default:
				continue
		}
	}
	return
}

func getTopFiveLogMessages(logMessages []LogMessage) (topFiveLogMessages []string, topFiveLogMessageFrequencies []int64) {
	rankedLogMessages := make(map[string]int64, len(logMessages))
	topFiveLogMessages = make([]string, 5)
	topFiveLogMessageFrequencies = make([]int64, 5)
	for _, logMessage := range logMessages {
		rankedLogMessages[logMessage.message] += 1
	}
	messages := make([]string, 0, len(rankedLogMessages))
	for message := range rankedLogMessages {
		messages = append(messages, message)
	}
	sort.SliceStable(messages, func(i, j int) bool{
		return rankedLogMessages[messages[i]] > rankedLogMessages[messages[j]]
	})
	if len(messages) == 0 {
		return
	}
	var maxMessages int
	if len(messages) >= 5 {
		maxMessages = 5
	} else {
		maxMessages = len(messages)
	}
	for index := 0; index < maxMessages; index++ {
		topFiveLogMessages[index] = messages[index]
		topFiveLogMessageFrequencies[index] = rankedLogMessages[messages[index]]
	}
	return
}

func getStartTime(logMessages []LogMessage) (startTime time.Time) {
	if len(logMessages) == 0 {
		return
	}
	startTime, err := time.Parse(layout, logMessages[0].timestamp)
	if err != nil {
		panic("Unable to parse start time")
	}
	return
}

func getEndTime(logMessages []LogMessage) (endTime time.Time) {
	if len(logMessages) == 0 {
		return
	}
	endTime, err := time.Parse(layout, logMessages[len(logMessages) - 1].timestamp)
	if err != nil {
		panic("Unable to parse end time")
	}
	return
}

func analyzeLogFile(logPath string, logAnalysisChan chan LogAnalysis) {
	logMessages := parseLogFile(logPath)
	var logAnalysis LogAnalysis
	logAnalysis.numEntries = getNumEntries(logMessages)
	logAnalysis.logSeverityFrequency = getLogSeverityFrequency(logMessages)
	logAnalysis.topFiveLogMessages, logAnalysis.topFiveLogMessageFrequencies = getTopFiveLogMessages(logMessages)
	logAnalysis.startTime = getStartTime(logMessages)
	logAnalysis.endTime = getEndTime(logMessages)
	logAnalysisChan <- logAnalysis	
	waitGroup.Done()
}

func printLogAnalysis(logAnalysis LogAnalysis) {
	fmt.Println("Number of Entries: " + strconv.Itoa(logAnalysis.numEntries))
	fmt.Println("Log Severity Frequency: ")
	fmt.Println("   DEBUG: " + strconv.FormatInt(logAnalysis.logSeverityFrequency.debug, 10))
	fmt.Println("   INFO: " + strconv.FormatInt(logAnalysis.logSeverityFrequency.info, 10))
	fmt.Println("   WARNING: " + strconv.FormatInt(logAnalysis.logSeverityFrequency.warning, 10))
	fmt.Println("   ERROR: " + strconv.FormatInt(logAnalysis.logSeverityFrequency.error, 10))
	fmt.Println("Top Five Log Messages: ")
	var maxMessages int
	if len(logAnalysis.topFiveLogMessages) >= 5 {
		maxMessages = 5
	} else {
		maxMessages = len(logAnalysis.topFiveLogMessages)
	}
	for index := 0; index < maxMessages; index ++ {
		fmt.Println("   " + strconv.Itoa(index + 1) + ". " + logAnalysis.topFiveLogMessages[index])
	}
	fmt.Println("Start Date/Time: " + logAnalysis.startTime.Format(layout))
	fmt.Println("End Date/Time: " + logAnalysis.endTime.Format(layout))
}

func analyzeTopFiveLogMessages(logAnalyses []LogAnalysis) (topFiveLogMessages []string) {
	rankedLogMessages := make(map[string]int64, len(logAnalyses))
	for _, logAnalysis := range logAnalyses {
		var maxMessages int
		if len(logAnalysis.topFiveLogMessages) >= 5 {
			maxMessages = 5
		} else {
			maxMessages = len(logAnalysis.topFiveLogMessages)
		}
		for index := 0; index < maxMessages; index ++ {
			rankedLogMessages[logAnalysis.topFiveLogMessages[index]] += logAnalysis.topFiveLogMessageFrequencies[index]
		}
	}
	
	// Sort the map of messages : frequency
	messages := make([]string, 0, len(logAnalyses))
	for message := range rankedLogMessages {
		messages = append(messages, message)
	}
	sort.SliceStable(messages, func(i, j int) bool{
		return rankedLogMessages[messages[i]] > rankedLogMessages[messages[j]]
	})
	var maxMessages int
	if len(messages) >= 5 {
		maxMessages = 5
	} else {
		maxMessages = len(messages)
	}
	for index := 0; index < maxMessages; index++ {
		topFiveLogMessages = append(topFiveLogMessages, messages[index])
	}
	fmt.Println(topFiveLogMessages)
	return	
}

func analyzelogAnalyses(logAnalyses []LogAnalysis) (finalLogAnalysis LogAnalysis) {
	if len(logAnalyses) == 0 {
		panic("No analysis found")
	}
	finalLogAnalysis.startTime = logAnalyses[0].startTime
	finalLogAnalysis.endTime = logAnalyses[0].endTime

	topFiveLogMessages := analyzeTopFiveLogMessages(logAnalyses)
	var maxMessages int
	if len(topFiveLogMessages) >= 5 {
		maxMessages = 5
	} else {
		maxMessages = len(topFiveLogMessages)
	}
	for index := 0; index < maxMessages; index ++ {
		finalLogAnalysis.topFiveLogMessages = append(finalLogAnalysis.topFiveLogMessages, topFiveLogMessages[index])
	}

	for _, logAnalysis := range logAnalyses {
		finalLogAnalysis.numEntries += logAnalysis.numEntries
		finalLogAnalysis.logSeverityFrequency.debug += logAnalysis.logSeverityFrequency.debug
		finalLogAnalysis.logSeverityFrequency.info += logAnalysis.logSeverityFrequency.info
		finalLogAnalysis.logSeverityFrequency.warning += logAnalysis.logSeverityFrequency.warning
		finalLogAnalysis.logSeverityFrequency.error += logAnalysis.logSeverityFrequency.error
		if finalLogAnalysis.startTime.After(logAnalysis.startTime) {
			finalLogAnalysis.startTime = logAnalysis.startTime
		}
		if finalLogAnalysis.endTime.Before(logAnalysis.endTime) {
			finalLogAnalysis.endTime = logAnalysis.endTime
		}
	}

	return
}

func analyzeLogFiles(logPaths []string) (logAnalysis LogAnalysis) {
	var logAnalysisChan chan LogAnalysis = make(chan LogAnalysis)
	var logAnalyses []LogAnalysis
	for _, logPath := range logPaths {
		waitGroup.Add(1)
		go analyzeLogFile(logPath, logAnalysisChan)
	}

	for range logPaths {
		logAnalysis := <- logAnalysisChan
		logAnalyses = append(logAnalyses, logAnalysis)
	}
	waitGroup.Wait()
	close(logAnalysisChan)
	logAnalysis = analyzelogAnalyses(logAnalyses)

	return
}

func main() {
	logPaths := os.Args[1:]
	logAnalysis := analyzeLogFiles(logPaths)
	printLogAnalysis(logAnalysis)
}