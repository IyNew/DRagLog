package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	// Add this import statement
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// the index defined for the couchdb
const index = "id~time"

type SimpleChaincode struct {
	contractapi.Contract
}

// type ReliabilityRecord struct {
// 	DataSourceID     string  `json:"dataSourceID"`
// 	ReliabilityScore float32 `json:"reliabilityScore"`
// 	Type             string  `json:"type"`
// 	Timestamp        string  `json:"timestamp"`
// 	Reserved         string  `json:"reserved"`
// }

type LogRecord struct {
	LogID            string  `json:"logID"`
	LoggerID         string  `json:"loggerID"`
	Type             string  `json:"type"`
	Input            string  `json:"input"`
	InputFrom        string  `json:"inputFrom"`
	Output           string  `json:"output"`
	OutputTo         string  `json:"outputTo"`
	ReliabilityScore float32 `json:"reliabilityScore"`
	Timestamp        string  `json:"timestamp"`
	Reserved         string  `json:"reserved"`
}

// Feedback is the feedback from the source to the user
type Feedback struct {
	Output   string `json:"sourceID"`
	Feedback string `json:"feedback"`
	Reserved string `json:"reserved"`
}

// HistoryLogRecord is the history log record
type HistoryLogRecord struct {
	Record    *LogRecord `json:"record"`
	Timestamp string     `json:"timestamp"`
	TxId      string     `json:"txID"`
	IsDelete  bool       `json:"isDelete"`
}

type PaginatedQueryResult struct {
	Records             []LogRecord `json:"records"`
	FetchedRecordsCount int32       `json:"fetchedRecordsCount"`
	Bookmark            string      `json:"bookmark"`
}

// Hello returns a greeting message to check if the chaincode is alive
func (s *SimpleChaincode) Hello(ctx contractapi.TransactionContextInterface) string {
	return "Hello from fabric, the service is running!"
}

// ReadReliabilityRecord returns the reliability record for the given data source ID
func (s *SimpleChaincode) ReadReliabilityRecord(ctx contractapi.TransactionContextInterface, dataSourceID string) (*LogRecord, error) {
	reliabilityRecordJSON, err := ctx.GetStub().GetState(dataSourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get the reliability record for the data source %s: %v", dataSourceID, err)
	}
	if reliabilityRecordJSON == nil {
		return nil, fmt.Errorf("the reliability record for the data source %s does not exist", dataSourceID)
	}

	var reliabilityRecord LogRecord
	err = json.Unmarshal(reliabilityRecordJSON, &reliabilityRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the reliability record for the data source %s: %v", dataSourceID, err)
	}

	return &reliabilityRecord, nil
}

// RecordExists returns true when record with given ID exists in world state
func (s *SimpleChaincode) RecordExists(ctx contractapi.TransactionContextInterface, recordID string) (bool, error) {
	recordJSON, err := ctx.GetStub().GetState(recordID)
	if err != nil {
		return false, fmt.Errorf("failed to read record %s from world state: %v", recordID, err)
	}

	return recordJSON != nil, nil
}

// ReadLogRecord returns the log record for the given log ID
func (s *SimpleChaincode) ReadLogRecord(ctx contractapi.TransactionContextInterface, logID string) (*LogRecord, error) {
	logRecordJSON, err := ctx.GetStub().GetState(logID)
	if err != nil {
		return nil, fmt.Errorf("failed to get the log record for the log ID %s: %v", logID, err)
	}
	if logRecordJSON == nil {
		return nil, fmt.Errorf("the log record for the log ID %s does not exist", logID)
	}

	var logRecord LogRecord
	err = json.Unmarshal(logRecordJSON, &logRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the log record for the log ID %s: %v", logID, err)
	}

	return &logRecord, nil
}

func (s *SimpleChaincode) ReadFeedbackRecord(ctx contractapi.TransactionContextInterface, logID string) (*LogRecord, error) {
	feedbackRecordJSON, err := ctx.GetStub().GetState(logID)
	if err != nil {
		return nil, fmt.Errorf("failed to get the feedback record for the log ID %s: %v", logID, err)
	}
	if feedbackRecordJSON == nil {
		return nil, fmt.Errorf("the feedback record for the log ID %s does not exist", logID)
	}

	var feedbackRecord LogRecord
	err = json.Unmarshal(feedbackRecordJSON, &feedbackRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the feedback record for the log ID %s: %v", logID, err)
	}

	return &feedbackRecord, nil
}

func MD5Hash(text string) string {
	hash := md5.New()
	hash.Write([]byte(text))
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *SimpleChaincode) CreateReliabilityRecord(ctx contractapi.TransactionContextInterface, dataSourceID string, digest string, reserved string) error {

	// check if the reliability record already exists
	exists, err := s.RecordExists(ctx, dataSourceID)
	if err != nil {
		return fmt.Errorf("failed to check if the reliability record for the data source %s exists: %v", dataSourceID, err)
	}
	if exists {
		return fmt.Errorf("the reliability record for the data source %s already exists", dataSourceID)
	}

	reliabilityRecord := LogRecord{
		LogID:            dataSourceID,
		LoggerID:         dataSourceID,
		Type:             "reliability",
		Input:            digest,
		InputFrom:        "",
		Output:           "",
		OutputTo:         "",
		ReliabilityScore: 100,
		Timestamp:        "0",
		Reserved:         reserved,
	}

	reliabilityRecordJSON, err := json.Marshal(reliabilityRecord)
	if err != nil {
		return fmt.Errorf("failed to marshal the reliability record for the data source %s: %v", dataSourceID, err)
	}

	err = ctx.GetStub().PutState(dataSourceID, reliabilityRecordJSON)
	if err != nil {
		return fmt.Errorf("failed to put the reliability record for the data source %s: %v", dataSourceID, err)
	}

	return nil
}

func (s *SimpleChaincode) CreateLogRecord(ctx contractapi.TransactionContextInterface, logID string, loggerID string, input string, inputFrom string, output string, outputTo string, timestamp string, reserved string) error {
	// check if the log record already exists with RecordExists
	exists, err := s.RecordExists(ctx, logID)
	if err != nil {
		return fmt.Errorf("failed to check if the log record for the log ID %s exists: %v", logID, err)
	}
	if exists {
		return fmt.Errorf("the log record for the log ID %s already exists", logID)
	}

	logRecord := LogRecord{
		LogID:            logID,
		LoggerID:         loggerID,
		Type:             "log",
		Input:            input,
		InputFrom:        inputFrom,
		Output:           output,
		OutputTo:         outputTo,
		Timestamp:        timestamp,
		ReliabilityScore: -1,
		Reserved:         reserved,
	}

	logRecordJSON, err := json.Marshal(logRecord)
	if err != nil {
		return fmt.Errorf("failed to marshal the log record for the log ID %s: %v", logID, err)
	}

	err = ctx.GetStub().PutState(logID, logRecordJSON)
	if err != nil {
		return fmt.Errorf("failed to put the log record for the log ID %s: %v", logID, err)
	}

	return nil
}

func (s *SimpleChaincode) CreateFeedbackRecord(ctx contractapi.TransactionContextInterface, logID string, loggerID string, input string, inputFrom string, output string, outputTo string, timestamp string, reserved string) error {
	// check if the feedback record already exists with RecordExists
	exists, err := s.RecordExists(ctx, logID)
	if err != nil {
		return fmt.Errorf("failed to check if the feedback record for the log ID %s exists: %v", logID, err)
	}
	if exists {
		return fmt.Errorf("the feedback record for the log ID %s already exists", logID)
	}

	fmt.Printf("logID: %s\n", logID)
	fmt.Printf("loggerID: %s\n", loggerID)
	fmt.Printf("input: %s\n", input)
	fmt.Printf("inputFrom: %s\n", inputFrom)
	fmt.Printf("output: %s\n", output)
	fmt.Printf("outputTo: %s\n", outputTo)
	fmt.Printf("timestamp: %s\n", timestamp)
	fmt.Printf("reserved: %s\n", reserved)

	feedbackRecord := LogRecord{
		LogID:            logID,
		LoggerID:         loggerID,
		Type:             "feedback",
		Input:            input,
		InputFrom:        inputFrom,
		Output:           output,
		OutputTo:         outputTo,
		Timestamp:        timestamp,
		ReliabilityScore: -1,
		Reserved:         reserved,
	}

	fmt.Printf("feedback record: %v\n", feedbackRecord)
	fmt.Printf("reserved field content: %s\n", reserved)

	feedbackRecordJSON, err := json.Marshal(feedbackRecord)
	if err != nil {
		return fmt.Errorf("failed to marshal the feedback record for the log ID %s: %v", logID, err)
	}

	err = ctx.GetStub().PutState(logID, feedbackRecordJSON)
	if err != nil {
		return fmt.Errorf("failed to put the feedback record for the log ID %s: %v", logID, err)
	}

	// parse the reserved field into a list of [id, score]
	// // example: reserved = "[[\"data/unreliable_large_source/sources_50.jsonl+916df390b8c436dc225ce59d63061245c28dfbc9545c064fef68d4721a59cdf8+1\", 0.001953125], [\"data/df_nq_valid_html_clean.jsonl+372dbac489b0fe04d85ae2cb23eafc294a0257c844705e9dd1a0fb7b82433fd3+2079\", 0.001953125], [\"data/df_nq_valid_html_clean.jsonl+2e8f02ca502f1d3724d0831fe6db2c662f6d722e14546a66bb1326098aa9d104+2478\", 0.001953125], [\"data/df_nq_valid_html_clean.jsonl+e95a31daefa032d5f6edf635926a5db5044b05ce4f2fe71f5fb743ee943a5b30+109\", 0.001953125]]"
	// type ScoreItem struct {
	// 	ID    string  `json:"0"`
	// 	Score float64 `json:"1"`
	// }
	// var reservedList []ScoreItem
	// err = json.Unmarshal([]byte(reserved), &reservedList)
	// if err != nil {
	// 	return fmt.Errorf("failed to unmarshal the reserved field for the log ID %s: %v", logID, err)
	// }

	// // update the reliability score of the data source
	// for _, item := range reservedList {
	// 	dataSourceID := item.ID
	// 	reliabilityScore := item.Score
	// 	reliabilityScore = -reliabilityScore * 10
	// 	err = s.UpdateReliabilityScore(ctx, dataSourceID, float32(reliabilityScore), true)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to update the reliability score for the data source %s: %v", dataSourceID, err)
	// 	}
	// 	fmt.Printf("updated the reliability score for the data source %s to %f\n", dataSourceID, reliabilityScore)
	// }

	return nil
}

// update the reliability score of the data source
func (s *SimpleChaincode) UpdateReliabilityScore(ctx contractapi.TransactionContextInterface, dataSourceID string, score float32, isDelta bool) error {
	reliabilityRecord, err := s.ReadReliabilityRecord(ctx, dataSourceID)
	if err != nil {
		return fmt.Errorf("failed to read the reliability record for the data source %s: %v", dataSourceID, err)
	}
	// print the current reliability record
	fmt.Printf("reliability record: %v\n", reliabilityRecord)

	if isDelta {
		reliabilityRecord.ReliabilityScore += score
	} else {
		reliabilityRecord.ReliabilityScore = score
	}

	reliabilityRecordJSON, err := json.Marshal(reliabilityRecord)
	if err != nil {
		return fmt.Errorf("failed to marshal the reliability record for the data source %s: %v", dataSourceID, err)
	}

	err = ctx.GetStub().PutState(dataSourceID, reliabilityRecordJSON)
	if err != nil {
		return fmt.Errorf("failed to put the reliability record for the data source %s: %v", dataSourceID, err)
	}

	return nil
}

// update the log record
func (s *SimpleChaincode) UpdateLogRecord(ctx contractapi.TransactionContextInterface, logID string, loggerID string, input string, inputFrom string, output string, outputTo string, timestamp string, reserved string) error {
	logRecord, err := s.ReadLogRecord(ctx, logID)
	if err != nil {
		return fmt.Errorf("failed to read the log record for the log ID %s: %v", logID, err)
	}

	logRecord.LoggerID = loggerID
	logRecord.Input = input
	logRecord.InputFrom = inputFrom
	logRecord.Output = output
	logRecord.OutputTo = outputTo
	logRecord.Timestamp = timestamp
	logRecord.Reserved = reserved

	logRecordJSON, err := json.Marshal(logRecord)
	if err != nil {
		return fmt.Errorf("failed to marshal the log record for the log ID %s: %v", logID, err)
	}

	err = ctx.GetStub().PutState(logID, logRecordJSON)
	if err != nil {
		return fmt.Errorf("failed to put the log record for the log ID %s: %v", logID, err)
	}

	return nil
}

// InitLedger adds the initial reliability record for the data source "default"
func (s *SimpleChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// err := s.CreateReliabilityRecord(ctx, "default", "default")
	// if err != nil {
	// 	return fmt.Errorf("failed to create the initial reliability record for the data source %s: %v", "default", err)
	// }
	//

	// create 10 reliability records with datasource id like "default0", "default1" ...
	for i := 0; i < 10; i++ {
		err := s.CreateReliabilityRecord(ctx, fmt.Sprintf("default%d", i), "default", "")
		if err != nil {
			return fmt.Errorf("failed to create the initial reliability record for the data source %s: %v", fmt.Sprintf("default%d", i), err)
		}
	}

	// create 10 log records with log id like "default0-reranker0", "default1-reranker0" ...
	for i := 0; i < 10; i++ {
		err := s.CreateLogRecord(ctx, fmt.Sprintf("default%d-reranker0", i), fmt.Sprintf("default%d", i), "", "", "default_output_from_datasource_default"+strconv.Itoa(i), "reranker0", "2025-01-01 00:00:00", "")
		if err != nil {
			return fmt.Errorf("failed to create the initial log record for the log ID %s: %v", fmt.Sprintf("default%d-reranker0", i), err)
		}
	}

	// create 1 log records with log id like reranker0-LLM0
	err := s.CreateLogRecord(ctx, "reranker0-LLM0", "reranker0", "", "", "reranker0_output_from_reranker0", "LLM0", "2025-01-02 00:00:00", "")
	if err != nil {
		return fmt.Errorf("failed to create the initial log record for the log ID %s: %v", "reranker0-LLM0", err)
	}

	return nil
}

// // GetLogRecord returns the log record for the given log ID
// func (s *SimpleChaincode) GetLogRecord(ctx contractapi.TransactionContextInterface, logID string) (*LogRecord, error) {
// 	logRecord, err := s.ReadLogRecord(ctx, logID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read the log record for the log ID %s: %v", logID, err)
// 	}
// 	return logRecord, nil
// }

// GetAllReliabilityRecords returns all reliability records found in world state
func (s *SimpleChaincode) GetAllRecords(ctx contractapi.TransactionContextInterface) ([]LogRecord, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []LogRecord
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var record LogRecord
		err = json.Unmarshal(queryResponse.Value, &record)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// constructQueryResponseFromIterator constructs a slices of Records from QueryResultsIterator
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*LogRecord, error) {
	var records []*LogRecord
	for resultsIterator.HasNext() {
		recordResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var record LogRecord
		err = json.Unmarshal(recordResponse.Value, &record)
		if err != nil {
			return nil, err
		}
		records = append(records, &record)
	}

	return records, nil
}

// getQueryResultForQueryString queries for records based on a passed in query string.
// This is only supported for couchdb
func (s *SimpleChaincode) getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*LogRecord, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// QueryRecords uses a query string to perform a query for records.
func (s *SimpleChaincode) QueryRecords(ctx contractapi.TransactionContextInterface, queryString string) ([]*LogRecord, error) {
	return s.getQueryResultForQueryString(ctx, queryString)
}

// QueryReliabilityRecords uses a query string to perform a query for reliability records.
func (s *SimpleChaincode) QueryReliabilityRecords(ctx contractapi.TransactionContextInterface, dataSourceID string) ([]*LogRecord, error) {
	queryString := fmt.Sprintf(`{"selector":{"LogID":"%s", "Type":"reliability"}}`, dataSourceID)
	return s.getQueryResultForQueryString(ctx, queryString)
}

// QueryLogRecords uses a query string to perform a query for log records.
func (s *SimpleChaincode) QueryLogRecords(ctx contractapi.TransactionContextInterface, logID string) ([]*LogRecord, error) {
	queryString := fmt.Sprintf(`{"selector":{"LogID":"%s", "Type":"log"}}`, logID)
	return s.getQueryResultForQueryString(ctx, queryString)
}

// QueryFeedbackRecords uses a query string to perform a query for feedback records.
func (s *SimpleChaincode) QueryFeedbackRecords(ctx contractapi.TransactionContextInterface, logID string) ([]*LogRecord, error) {
	queryString := fmt.Sprintf(`{"selector":{"LogID":"%s", "Type":"feedback"}}`, logID)
	return s.getQueryResultForQueryString(ctx, queryString)
}

// GetHistoryForRecord returns the history of a record for a given record ID.
func (s *SimpleChaincode) GetHistoryForRecord(ctx contractapi.TransactionContextInterface, recordID string) ([]HistoryLogRecord, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(recordID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var historyLogs []HistoryLogRecord
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var logRecord LogRecord
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &logRecord)
			if err != nil {
				return nil, err
			}
		} else {
			logRecord = LogRecord{
				LogID: recordID,
			}
		}

		historyLog := HistoryLogRecord{
			Record:    &logRecord,
			Timestamp: fmt.Sprintf("%d.%09d", response.Timestamp.Seconds, response.Timestamp.Nanos),
			TxId:      response.TxId,
			IsDelete:  response.IsDelete,
		}

		historyLogs = append(historyLogs, historyLog)
	}

	return historyLogs, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SimpleChaincode{})
	if err != nil {
		log.Panicf("Error creating asset chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting asset chaincode: %v", err)
	}
}
