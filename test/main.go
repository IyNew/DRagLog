package main

import (
	"context"
	"draglog_api/utils"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
)

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

type LogRecordResponse struct {
	Body struct {
		Message string      `json:"message" doc:"Response message"`
		Records []LogRecord `json:"records" doc:"List of log records"`
	}
}

type Options struct {
	Port int `help:"Port to listen on" short:"p" default:"8080"`
}

func main() {
	utils.InitGateway()

	//
	// utils.basicInitLedger()
	// utils.basicCreateLogRecord()
	// utils.basicGetAllLogRecords()

	// create a huma cli app which takes a port option
	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		// Create a new router & API
		router := chi.NewMux()
		api := humachi.New(router, huma.DefaultConfig("My API", "1.0.0"))

		// Register GET /init-ledger
		huma.Register(api, huma.Operation{
			OperationID: "basicInitLedger",
			Method:      http.MethodGet,
			Path:        "/init-ledger",
			Summary:     "Init the ledger",
			Description: "Init the ledger",
			Tags:        []string{"Init"},
		}, func(ctx context.Context, input *struct{}) (*struct{}, error) {
			utils.basicInitLedger()
			return &struct{}{}, nil
		})

		// Register GET /basicCreateLogRecord
		huma.Register(api, huma.Operation{
			OperationID: "basicCreateLogRecord",
			Method:      http.MethodGet,
			Path:        "/create-log-record",
			Summary:     "Create a log record",
			Description: "Create a log record",
			Tags:        []string{"Create"},
		}, func(ctx context.Context, input *struct{}) (*struct{}, error) {
			utils.basicCreateLogRecord()
			return &struct{}{}, nil
		})

		huma.Register(api, huma.Operation{
			OperationID: "basicGetAllLogRecords",
			Method:      http.MethodGet,
			Path:        "/get-all-log-records",
			Summary:     "Get all log records",
			Description: "Get all log records",
			Tags:        []string{"Get"},
		}, func(ctx context.Context, input *struct{}) (*LogRecordResponse, error) {
			result := utils.basicGetAllLogRecords()

			// Parse the JSON result into LogRecord slice
			var records []LogRecord
			if err := json.Unmarshal([]byte(result), &records); err != nil {
				return nil, fmt.Errorf("failed to parse log records: %w", err)
			}

			resp := &LogRecordResponse{}
			resp.Body.Message = fmt.Sprintf("Found %d log records", len(records))
			resp.Body.Records = records
			return resp, nil
		})

		// Start the server
		hooks.OnStart(func() error {
			fmt.Printf("Starting server on port %d...\n", options.Port)
			http.ListenAndServe(fmt.Sprintf(":%d", options.Port), router)
		})
	})

	// Run the CLI
	cli.Run()
}
