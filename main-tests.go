package main

import (
	"os"
	"testing"
	"time"
	"reflect"
)

func TestParseLogMessage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    LogMessage
		wantErr bool
	}{
		{
			name:  "valid log message",
			input: "2024-01-02 15:04:05.999 | INFO | app.module: function: 123 - User logged in",
			want: LogMessage{
				timestamp:   "2024-01-02 15:04:05.999",
				severity:    "INFO",
				module:     "app.module",
				function:   "function",
				lineNumber: 123,
				message:    "User logged in",
			},
			wantErr: false,
		},
		{
			name:    "empty message",
			input:   "",
			want:    LogMessage{},
			wantErr: true,
		},
		{
			name:    "malformed message - missing severity",
			input:   "2024-01-02 15:04:05.999 | | app.module: function: 123 - User logged in",
			want:    LogMessage{},
			wantErr: true,
		},
		{
			name:    "malformed message - missing line number",
			input:   "2024-01-02 15:04:05.999 | INFO | app.module: function: - User logged in",
			want:    LogMessage{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLogMessage(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLogMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLogMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLogSeverityFrequency(t *testing.T) {
	testLogs := []LogMessage{
		{severity: "DEBUG"},
		{severity: "INFO"},
		{severity: "INFO"},
		{severity: "WARNING"},
		{severity: "ERROR"},
		{severity: "ERROR"},
		{severity: "INVALID"},
	}

	want := LogSeverityFrequency{
		debug:   1,
		info:    2,
		warning: 1,
		error:   2,
	}

	got := getLogSeverityFrequency(testLogs)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("getLogSeverityFrequency() = %v, want %v", got, want)
	}
}

func TestGetTopFiveLogMessages(t *testing.T) {
	testLogs := []LogMessage{
		{message: "Error 1"},
		{message: "Error 1"},
		{message: "Error 2"},
		{message: "Error 3"},
		{message: "Error 3"},
		{message: "Error 3"},
		{message: "Error 4"},
		{message: "Error 5"},
		{message: "Error 6"},
	}

	wantMessages := []string{"Error 3", "Error 1", "Error 2", "Error 4", "Error 5"}
	wantFrequencies := []int64{3, 2, 1, 1, 1}

	gotMessages, gotFrequencies := getTopFiveLogMessages(testLogs)
	
	if !reflect.DeepEqual(gotMessages, wantMessages) {
		t.Errorf("getTopFiveLogMessages() messages = %v, want %v", gotMessages, wantMessages)
	}
	if !reflect.DeepEqual(gotFrequencies, wantFrequencies) {
		t.Errorf("getTopFiveLogMessages() frequencies = %v, want %v", gotFrequencies, wantFrequencies)
	}
}

func TestGetStartAndEndTime(t *testing.T) {
	testLogs := []LogMessage{
		{timestamp: "2024-01-01 00:00:00.000"},
		{timestamp: "2024-01-01 12:00:00.000"},
		{timestamp: "2024-01-02 00:00:00.000"},
	}

	expectedStart, _ := time.Parse(layout, "2024-01-01 00:00:00.000")
	expectedEnd, _ := time.Parse(layout, "2024-01-02 00:00:00.000")

	gotStart := getStartTime(testLogs)
	gotEnd := getEndTime(testLogs)

	if !gotStart.Equal(expectedStart) {
		t.Errorf("getStartTime() = %v, want %v", gotStart, expectedStart)
	}
	if !gotEnd.Equal(expectedEnd) {
		t.Errorf("getEndTime() = %v, want %v", gotEnd, expectedEnd)
	}
}

// Helper function to create temporary test log file
func createTestLogFile(t *testing.T, content string) string {
	tmpfile, err := os.CreateTemp("", "test-log-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	
	return tmpfile.Name()
}

func TestAnalyzeLogFile(t *testing.T) {
	logContent := `2024-01-01 00:00:00.000 | INFO | app.module: function: 123 - User logged in
2024-01-01 00:01:00.000 | ERROR | app.module: function: 124 - Database connection failed
2024-01-01 00:02:00.000 | ERROR | app.module: function: 125 - Database connection failed`

	tmpFileName := createTestLogFile(t, logContent)
	defer os.Remove(tmpFileName)

	logAnalysisChan := make(chan LogAnalysis)
	waitGroup.Add(1)
	
	go analyzeLogFile(tmpFileName, logAnalysisChan)
	
	logAnalysis := <-logAnalysisChan
	waitGroup.Wait()

	if logAnalysis.numEntries != 3 {
		t.Errorf("Expected 3 entries, got %d", logAnalysis.numEntries)
	}
	
	if logAnalysis.logSeverityFrequency.info != 1 || logAnalysis.logSeverityFrequency.error != 2 {
		t.Errorf("Incorrect severity frequencies: got info=%d, error=%d, want info=1, error=2",
			logAnalysis.logSeverityFrequency.info, logAnalysis.logSeverityFrequency.error)
	}

	expectedMessage := "Database connection failed"
	if logAnalysis.topFiveLogMessages[0] != expectedMessage {
		t.Errorf("Expected top message to be '%s', got '%s'", 
			expectedMessage, logAnalysis.topFiveLogMessages[0])
	}
}

func TestAnalyzeLogFiles(t *testing.T) {
	log1Content := `2024-01-01 00:00:00.000 | INFO | app.module: function: 123 - User logged in
2024-01-01 00:01:00.000 | ERROR | app.module: function: 124 - Database error`

	log2Content := `2024-01-01 00:02:00.000 | WARNING | app.module: function: 125 - Low memory
2024-01-01 00:03:00.000 | ERROR | app.module: function: 126 - Database error`

	tmpFile1 := createTestLogFile(t, log1Content)
	tmpFile2 := createTestLogFile(t, log2Content)
	defer os.Remove(tmpFile1)
	defer os.Remove(tmpFile2)

	logPaths := []string{tmpFile1, tmpFile2}
	analysis := analyzeLogFiles(logPaths)

	// Test basic metrics
	if analysis.numEntries != 4 {
		t.Errorf("Expected 4 total entries, got %d", analysis.numEntries)
	}

	// Test severity frequencies
	expectedFreq := LogSeverityFrequency{
		info:    1,
		warning: 1,
		error:   2,
	}
	if !reflect.DeepEqual(analysis.logSeverityFrequency, expectedFreq) {
		t.Errorf("Incorrect severity frequencies: got %+v, want %+v",
			analysis.logSeverityFrequency, expectedFreq)
	}

	// Test top message
	expectedTopMessage := "Database error"
	if analysis.topFiveLogMessages[0] != expectedTopMessage {
		t.Errorf("Expected top message to be '%s', got '%s'",
			expectedTopMessage, analysis.topFiveLogMessages[0])
	}
}
