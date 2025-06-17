package main

import (
	"context"
	"draglog_api/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
)

type LogRecord struct {
	LogID            string  `json:"logID" default:"default0-reranker0"`
	LoggerID         string  `json:"loggerID" default:"reranker"`
	Type             string  `json:"type" default:"log"`
	Input            string  `json:"input" default:"test_input"`
	InputFrom        string  `json:"inputFrom" default:"test_input_from"`
	Output           string  `json:"output" default:"test_output"`
	OutputTo         string  `json:"outputTo" default:"test_output_to"`
	ReliabilityScore float32 `json:"reliabilityScore" default:"-1"`
	Timestamp        string  `json:"timestamp" default:"test_timestamp"`
	Reserved         string  `json:"reserved" default:"test_reserved"`
}

type LogRecordResponse struct {
	Body struct {
		Message string      `json:"message" doc:"Response message"`
		Records []LogRecord `json:"records" doc:"List of log records"`
	}
}

type LogRecordHistory struct {
	Record    *LogRecord `json:"record"`
	Timestamp string     `json:"timestamp"`
	TxId      string     `json:"txID"`
	IsDelete  bool       `json:"isDelete"`
}

type LogRecordHistoryResponse struct {
	Body struct {
		Message string             `json:"message" doc:"Response message"`
		History []LogRecordHistory `json:"history" doc:"List of log record histories"`
	}
}

type Selector struct {
	Body struct {
		Selector string `json:"selector" doc:"Selector" default:"{\"selector\": {\"type\": \"log\"}}"`
	}
}

type Options struct {
	Port int `help:"Port to listen on" short:"p" default:"8080"`
}

var debug = true
var debugLogFile *os.File

func initDebugLog() error {
	if !debug {
		return nil
	}

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Open log file with append mode
	var err error
	debugLogFile, err = os.OpenFile("logs/api_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open debug log file: %w", err)
	}
	return nil
}

func logDebugData(operation string, data interface{}) error {
	if !debug || debugLogFile == nil {
		return nil
	}

	logEntry := struct {
		Timestamp string      `json:"timestamp"`
		Operation string      `json:"operation"`
		Data      interface{} `json:"data"`
	}{
		Timestamp: time.Now().Format(time.RFC3339),
		Operation: operation,
		Data:      data,
	}

	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal debug data: %w", err)
	}

	if _, err := debugLogFile.Write(append(jsonData, '\n')); err != nil {
		return fmt.Errorf("failed to write to debug log: %w", err)
	}

	return nil
}

func main() {
	utils.InitGateway()

	// Initialize debug logging if enabled
	if err := initDebugLog(); err != nil {
		fmt.Printf("Warning: Failed to initialize debug logging: %v\n", err)
	}

	// create a huma cli app which takes a port option
	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		// Create a new router & API
		router := chi.NewMux()
		api := humachi.New(router, huma.DefaultConfig("My API", "1.0.0"))

		// Register GET /init-ledger
		huma.Register(api, huma.Operation{
			OperationID: "initLedger",
			Method:      http.MethodGet,
			Path:        "/init-ledger",
			Summary:     "Init the ledger",
			Description: "Init the ledger",
			Tags:        []string{"Init"},
		}, func(ctx context.Context, input *struct{}) (*struct{}, error) {
			utils.InitLedgerTest()
			return &struct{}{}, nil
		})

		// Register GET /CreateLogRecord
		huma.Register(api, huma.Operation{
			OperationID: "CreateLogRecord",
			Method:      http.MethodPost,
			Path:        "/create-log-record",
			Summary:     "Create a log record",
			Description: "Create a new log record with the provided details",
			Tags:        []string{"Create"},
		}, func(ctx context.Context, input *struct {
			Body LogRecord `json:"body" doc:"Log record details"`
		}) (*struct{}, error) {
			if err := logDebugData("create-log-record", input.Body); err != nil {
				fmt.Printf("Warning: Failed to log debug data: %v\n", err)
			}
			utils.CreateLogRecord(
				input.Body.LogID,
				input.Body.LoggerID,
				input.Body.Input,
				input.Body.InputFrom,
				input.Body.Output,
				input.Body.OutputTo,
				input.Body.Timestamp,
				input.Body.Reserved,
			)
			return &struct{}{}, nil
		})

		// Register POST /create-feedback-record
		huma.Register(api, huma.Operation{
			OperationID: "CreateFeedbackRecord",
			Method:      http.MethodPost,
			Path:        "/create-feedback-record",
			Summary:     "Create a feedback record",
			Description: "Create a new feedback record with the provided details",
			Tags:        []string{"Create"},
		}, func(ctx context.Context, input *struct {
			Body LogRecord `json:"body" doc:"Log record details"`
		}) (*struct{}, error) {
			if err := logDebugData("create-feedback-record", input.Body); err != nil {
				fmt.Printf("Warning: Failed to log debug data: %v\n", err)
			}
			utils.CreateFeedbackRecord(
				input.Body.LogID,
				input.Body.LoggerID,
				input.Body.Input,
				input.Body.InputFrom,
				input.Body.Output,
				input.Body.OutputTo,
				input.Body.Timestamp,
				input.Body.Reserved,
			)
			return &struct{}{}, nil
		})

		// Register POST /create-reliability-record
		huma.Register(api, huma.Operation{
			OperationID: "CreateReliabilityRecord",
			Method:      http.MethodPost,
			Path:        "/create-reliability-record",
			Summary:     "Create a reliability record",
			Description: "Create a new reliability record",
			Tags:        []string{"Create"},
		}, func(ctx context.Context, input *struct {
			Body struct {
				DataSourceID string `json:"dataSourceID" doc:"Data source ID"`
				Digest       string `json:"digest" doc:"Digest value"`
				Reserved     string `json:"reserved" doc:"Reserved value"`
			}
		}) (*struct{}, error) {
			if err := logDebugData("create-reliability-record", input.Body); err != nil {
				fmt.Printf("Warning: Failed to log debug data: %v\n", err)
			}
			utils.CreateReliabilityRecord(input.Body.DataSourceID, input.Body.Digest, input.Body.Reserved)
			return &struct{}{}, nil
		})

		// Register POST /create-reliability-records-batch
		huma.Register(api, huma.Operation{
			OperationID: "CreateReliabilityRecordsBatch",
			Method:      http.MethodPost,
			Path:        "/create-reliability-records-batch",
			Summary:     "Create reliability records in batch",
		}, func(ctx context.Context, input *struct {
			Body struct {
				RecordsJSON string `json:"recordsJSON" doc:"JSON string of log records"`
			}
		}) (*struct{}, error) {
			if err := logDebugData("create-reliability-records-batch", input.Body); err != nil {
				fmt.Printf("Warning: Failed to log debug data: %v\n", err)
			}
			utils.CreateReliabilityRecordsBatch(input.Body.RecordsJSON)
			return &struct{}{}, nil
		})

		// Register POST /create-reliability-record-async
		huma.Register(api, huma.Operation{
			OperationID: "CreateReliabilityRecordAsync",
			Method:      http.MethodPost,
			Path:        "/create-reliability-record-async",
			Summary:     "Create a reliability record asynchronously",
		}, func(ctx context.Context, input *struct {
			Body struct {
				DataSourceID string `json:"dataSourceID" doc:"Data source ID"`
				Digest       string `json:"digest" doc:"Digest value"`
				Reserved     string `json:"reserved" doc:"Reserved value"`
			}
		}) (*struct{}, error) {
			if err := logDebugData("create-reliability-record-async", input.Body); err != nil {
				fmt.Printf("Warning: Failed to log debug data: %v\n", err)
			}
			utils.CreateReliabilityRecordAsync(input.Body.DataSourceID, input.Body.Digest, input.Body.Reserved)
			return &struct{}{}, nil
		})

		// Register GET /get-all-log-records
		huma.Register(api, huma.Operation{
			OperationID: "GetAllLogRecords",
			Method:      http.MethodGet,
			Path:        "/get-all-log-records",
			Summary:     "Get all log records",
			Description: "Get all log records from the ledger",
			Tags:        []string{"Get"},
		}, func(ctx context.Context, input *struct{}) (*LogRecordResponse, error) {
			result := utils.GetAllLogRecords()
			var records []LogRecord
			if err := json.Unmarshal([]byte(result), &records); err != nil {
				return nil, fmt.Errorf("failed to parse log records: %w", err)
			}
			resp := &LogRecordResponse{}
			resp.Body.Message = fmt.Sprintf("Found %d log records", len(records))
			resp.Body.Records = records
			return resp, nil
		})

		// Register GET /get-all-reliability-records
		huma.Register(api, huma.Operation{
			OperationID: "GetAllReliabilityRecords",
			Method:      http.MethodGet,
			Path:        "/get-all-reliability-records",
			Summary:     "Get all reliability records",
			Description: "Get all reliability records from the ledger",
			Tags:        []string{"Get"},
		}, func(ctx context.Context, input *struct{}) (*LogRecordResponse, error) {
			result := utils.GetAllReliabilityRecords()
			var records []LogRecord
			if err := json.Unmarshal([]byte(result), &records); err != nil {
				return nil, fmt.Errorf("failed to parse reliability records: %w", err)
			}
			resp := &LogRecordResponse{}
			resp.Body.Message = fmt.Sprintf("Found %d reliability records", len(records))
			resp.Body.Records = records
			return resp, nil
		})

		// Register GET /get-all-feedback-records
		huma.Register(api, huma.Operation{
			OperationID: "GetAllFeedbackRecords",
			Method:      http.MethodGet,
			Path:        "/get-all-feedback-records",
			Summary:     "Get all feedback records",
			Description: "Get all feedback records from the ledger",
			Tags:        []string{"Get"},
		}, func(ctx context.Context, input *struct{}) (*LogRecordResponse, error) {
			result := utils.GetAllFeedbackRecords()
			var records []LogRecord
			if err := json.Unmarshal([]byte(result), &records); err != nil {
				return nil, fmt.Errorf("failed to parse feedback records: %w", err)
			}
			resp := &LogRecordResponse{}
			resp.Body.Message = fmt.Sprintf("Found %d feedback records", len(records))
			resp.Body.Records = records
			return resp, nil
		})

		// Register GET /get-log-record/{logID}
		huma.Register(api, huma.Operation{
			OperationID: "GetLogRecord",
			Method:      http.MethodGet,
			Path:        "/get-log-record/{logID}",
			Summary:     "Get a log record",
			Description: "Get a specific log record by ID",
			Tags:        []string{"Get"},
		}, func(ctx context.Context, input *struct {
			LogID string `path:"logID" doc:"Log record ID"`
		}) (*LogRecordResponse, error) {
			result := utils.GetLogRecord(input.LogID)
			var record LogRecord
			if err := json.Unmarshal([]byte(result), &record); err != nil {
				return nil, fmt.Errorf("failed to parse log record: %w", err)
			}
			resp := &LogRecordResponse{}
			resp.Body.Message = "Found log record"
			resp.Body.Records = []LogRecord{record}
			return resp, nil
		})

		// Register GET /get-reliability-record/{dataSourceID}
		huma.Register(api, huma.Operation{
			OperationID: "GetReliabilityRecord",
			Method:      http.MethodGet,
			Path:        "/get-reliability-record/{dataSourceID}",
			Summary:     "Get a reliability record",
			Description: "Get a specific reliability record by data source ID",
			Tags:        []string{"Get"},
		}, func(ctx context.Context, input *struct {
			DataSourceID string `path:"dataSourceID" doc:"Data source ID"`
		}) (*LogRecordResponse, error) {
			result := utils.GetReliabilityRecord(input.DataSourceID)
			var record LogRecord
			if err := json.Unmarshal([]byte(result), &record); err != nil {
				return nil, fmt.Errorf("failed to parse reliability record: %w", err)
			}
			resp := &LogRecordResponse{}
			resp.Body.Message = "Found reliability record"
			resp.Body.Records = []LogRecord{record}
			return resp, nil
		})

		// Register PUT /update-reliability-record/{dataSourceID}
		huma.Register(api, huma.Operation{
			OperationID: "UpdateReliabilityRecord",
			Method:      http.MethodPut,
			Path:        "/update-reliability-record/{dataSourceID}",
			Summary:     "Update a reliability record",
			Description: "Update the reliability score for a specific data source",
			Tags:        []string{"Update"},
		}, func(ctx context.Context, input *struct {
			DataSourceID string `path:"dataSourceID" doc:"Data source ID"`
			Body         struct {
				ReliabilityScore float32 `json:"reliabilityScore" doc:"New reliability score"`
				IsDelta          bool    `json:"isDelta" doc:"Is delta"`
				Info             string  `json:"info" doc:"Info"`
			}
		}) (*struct{}, error) {
			utils.UpdateReliabilityRecord(input.DataSourceID, input.Body.ReliabilityScore, input.Body.IsDelta, input.Body.Info)
			return &struct{}{}, nil
		})

		// Register GET /get-feedback-record/{logID}
		huma.Register(api, huma.Operation{
			OperationID: "GetFeedbackRecord",
			Method:      http.MethodGet,
			Path:        "/get-feedback-record/{logID}",
			Summary:     "Get a feedback record",
		}, func(ctx context.Context, input *struct {
			LogID string `path:"logID" doc:"Log record ID"`
		}) (*LogRecordResponse, error) {
			result := utils.GetFeedbackRecord(input.LogID)
			var record LogRecord
			if err := json.Unmarshal([]byte(result), &record); err != nil {
				return nil, fmt.Errorf("failed to parse feedback record: %w", err)
			}
			resp := &LogRecordResponse{}
			resp.Body.Message = "Found feedback record"
			resp.Body.Records = []LogRecord{record}
			return resp, nil
		})

		// Register GET /get-record-with-selector
		huma.Register(api, huma.Operation{
			OperationID: "GetRecordWithSelector",
			Method:      http.MethodGet,
			Path:        "/get-record-with-selector",
			Summary:     "Get a record with a selector",
		}, func(ctx context.Context, input *struct {
			Body struct {
				Selector string `json:"selector" doc:"Selector" default:"{\"selector\": {\"type\": \"log\"}}"`
			}
		}) (*LogRecordResponse, error) {
			result := utils.GetRecordWithSelector(input.Body.Selector)
			var records []LogRecord
			if err := json.Unmarshal([]byte(result), &records); err != nil {
				return nil, fmt.Errorf("failed to parse records: %w", err)
			}
			resp := &LogRecordResponse{}
			resp.Body.Message = fmt.Sprintf("Found %d records", len(records))
			resp.Body.Records = records
			return resp, nil
		})

		// Register GET /get-history-for-record/{logID}
		huma.Register(api, huma.Operation{
			OperationID: "GetHistoryForRecord",
			Method:      http.MethodGet,
			Path:        "/get-history-for-record/{logID}",
			Summary:     "Get record history",
			Description: "Get the history of changes for a specific record",
			Tags:        []string{"Get"},
		}, func(ctx context.Context, input *struct {
			LogID string `path:"logID" doc:"Log record ID"`
		}) (*LogRecordHistoryResponse, error) {
			result := utils.GetHistoryForRecord(input.LogID)
			var history []LogRecordHistory
			if err := json.Unmarshal([]byte(result), &history); err != nil {
				return nil, fmt.Errorf("failed to parse record history: %w", err)
			}
			resp := &LogRecordHistoryResponse{}
			resp.Body.Message = fmt.Sprintf("Found %d history records", len(history))
			resp.Body.History = history
			return resp, nil
		})

		// Start the server
		hooks.OnStart(func() {
			fmt.Printf("Starting server on port %d...\n", options.Port)
			http.ListenAndServe(fmt.Sprintf(":%d", options.Port), router)
		})
	})

	// Run the CLI
	cli.Run()
}
