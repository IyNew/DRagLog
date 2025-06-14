package utils

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	mspID        = "Org1MSP"
	cryptoPath   = "../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com"
	certPath     = cryptoPath + "/users/User1@org1.example.com/msp/signcerts"
	keyPath      = cryptoPath + "/users/User1@org1.example.com/msp/keystore"
	tlsCertPath  = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint = "dns:///localhost:7051"
	gatewayPeer  = "peer0.org1.example.com"
)

var (
	GatewayConn    *client.Gateway
	ClientConn     *grpc.ClientConn
	ClientContract *client.Contract
)

// testing
var now = time.Now()
var assetId = fmt.Sprintf("asset%d", now.Unix()*1e3+int64(now.Nanosecond())/1e6)

// InitGateway initializes the Gateway connection.
func InitGateway() {
	ClientConn = newGrpcConnection()

	id := newIdentity()
	sign := newSign()
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(ClientConn),
		client.WithEvaluateTimeout(10*time.Second),
		client.WithEndorseTimeout(30*time.Second),
		client.WithSubmitTimeout(10*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	GatewayConn = gw

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := "basic"
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	network := gw.GetNetwork(channelName)
	ClientContract = network.GetContract(chaincodeName)
}

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() *grpc.ClientConn {
	certificatePEM, err := os.ReadFile(tlsCertPath)
	if err != nil {
		panic(fmt.Errorf("failed to read TLS certifcate file: %w", err))
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.NewClient(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() *identity.X509Identity {
	certificatePEM, err := readFirstFile(certPath)
	if err != nil {
		panic(fmt.Errorf("failed to read certificate file: %w", err))
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign() identity.Sign {
	privateKeyPEM, err := readFirstFile(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}

func readFirstFile(dirPath string) ([]byte, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}

	fileNames, err := dir.Readdirnames(1)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(path.Join(dirPath, fileNames[0]))
}

// Format JSON data
func formatJSON(data []byte) string {
	// if the data is empty, return an empty string
	if len(data) == 0 {
		return "[]"
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()
}

func InitLedgerTest() {
	fmt.Printf("\n--> Submit Transaction: InitLedger, function creates the initial set of log records on the ledger \n")

	_, err := ClientContract.SubmitTransaction("InitLedger")
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Printf("*** Transaction committed successfully\n")
}

// Evaluate a transaction to query ledger state.
func testGetAllLogRecord() {
	fmt.Println("\n--> Evaluate Transaction: GetAllLogRecord, function returns all the current log records on the ledger")

	evaluateResult, err := ClientContract.EvaluateTransaction("GetAllRecords")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("*** Result:%s\n", result)
}

// Submit a transaction synchronously, blocking until it has been committed to the ledger.
func testCreateLogRecord() {
	fmt.Printf("\n--> Submit Transaction: CreateLogRecord, creates new log record with logID, loggerID, input, inputFrom, output, outputTo, timestamp and reserved arguments \n")

	_, err := ClientContract.SubmitTransaction("CreateLogRecord", "test_log_id", "test_logger_id", "test_input", "test_input_from", "test_output", "test_output_to", "test_timestamp", "test_reserved")
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Printf("*** Transaction committed successfully\n")
}

func testGetAllReliabilityRecords() {
	fmt.Println("\n--> Evaluate Transaction: GetAllReliabilityRecords, function returns all the current reliability records on the ledger")

	evaluateResult, err := ClientContract.EvaluateTransaction("GetAllReliabilityRecords")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("*** Result:%s\n", result)
}

func testCreateReliabilityRecord() {
	fmt.Printf("\n--> Submit Transaction: CreateReliabilityRecord, creates new reliability record with datasourceID and test_digest\n")

	_, err := ClientContract.SubmitTransaction("CreateReliabilityRecord", "test_data_source_id", "test_digest")
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Printf("*** Transaction committed successfully\n")
}

func CreateLogRecord(logID string, loggerID string, input string, inputFrom string, output string, outputTo string, timestamp string, reserved string) {
	_, err := ClientContract.SubmitTransaction("CreateLogRecord", logID, loggerID, input, inputFrom, output, outputTo, timestamp, reserved)
	if err != nil {
		fmt.Printf("failed to submit transaction: %v\n", err)
		return
	}

	fmt.Printf("*** Transaction committed successfully\n")
}

func CreateReliabilityRecord(dataSourceID string, digest string, reserved string) {
	_, err := ClientContract.SubmitTransaction("CreateReliabilityRecord", dataSourceID, digest, reserved)
	if err != nil {
		fmt.Printf("failed to submit transaction: %v\n", err)
		return
	}
}

func CreateFeedbackRecord(logID string, loggerID string, input string, inputFrom string, output string, outputTo string, timestamp string, reserved string) {
	_, err := ClientContract.SubmitTransaction("CreateFeedbackRecord", logID, loggerID, input, inputFrom, output, outputTo, timestamp, reserved)
	if err != nil {
		fmt.Printf("failed to submit transaction: %v\n", err)
		return
	}
}

// func CreateLogRecordAsync(logID string, loggerID string, input string, inputFrom string, output string, outputTo string, timestamp string, reserved string) {
// 	// fmt.Printf("\n--> Submit Transaction: CreateRecord, creates new record with droneID, zip, flytime, flyrecord and reserved arguments \n")

// 	submitResult, commit, err := ClientContract.SubmitAsync("CreateLogRecord", client.WithArguments(logID, loggerID, input, inputFrom, output, outputTo, timestamp, reserved))
// 	if err != nil {
// 		// panic(fmt.Errorf("failed to submit transaction asynchronously: %w", err))
// 		fmt.Printf("failed to submit transaction asynchronously: %v\n", err)
// 		// fmt.Printf("Please check if the record %s, %s, %s already exists\n", logID, loggerID, input)
// 		return
// 	}

// 	fmt.Printf("\n*** Successfully submitted transaction to store the record: %s_%s. Info: %s\n", logID, timestamp, string(submitResult))
// 	// fmt.Println("*** Waiting for transaction commit.")

// 	if commitStatus, err := commit.Status(); err != nil {
// 		panic(fmt.Errorf("failed to get commit status: %w", err))
// 	} else if !commitStatus.Successful {
// 		panic(fmt.Errorf("transaction %s failed to commit with status: %d", commitStatus.TransactionID, int32(commitStatus.Code)))
// 	}

// 	// fmt.Printf("*** Transaction committed successfully\n")
// }

func GetAllLogRecords() string {
	// fmt.Println("\n--> Evaluate Transaction: GetAllRecords, function returns all the current records on the ledger")

	selector := `{"selector": {"type": "log"}}`
	return GetRecordWithSelector(selector)
}

func GetAllReliabilityRecords() string {
	selector := `{"selector": {"type": "reliability"}}`
	return GetRecordWithSelector(selector)
}

func GetAllFeedbackRecords() string {
	selector := `{"selector": {"type": "feedback"}}`
	return GetRecordWithSelector(selector)
}

func GetLogRecord(logID string) string {
	evaluateResult, err := ClientContract.EvaluateTransaction("ReadLogRecord", logID)
	if err != nil {
		fmt.Printf("failed to evaluate transaction: %v\n", err)
		return ""
	}
	result := formatJSON(evaluateResult)
	return result
}

func GetReliabilityRecord(dataSourceID string) string {
	evaluateResult, err := ClientContract.EvaluateTransaction("ReadReliabilityRecord", dataSourceID)
	if err != nil {
		fmt.Printf("failed to evaluate transaction: %v\n", err)
		return ""
	}
	result := formatJSON(evaluateResult)
	return result
}

func GetFeedbackRecord(logID string) string {
	evaluateResult, err := ClientContract.EvaluateTransaction("ReadFeedbackRecord", logID)
	if err != nil {
		fmt.Printf("failed to evaluate transaction: %v\n", err)
		return ""
	}
	result := formatJSON(evaluateResult)
	return result
}

// func getAllLogRecords() string {
// 	selector := `{"selector": {"type": "log"}}`
// 	return getRecordWithSelector(selector)
// }

func GetRecordWithSelector(selector string) string {
	evaluateResult, err := ClientContract.EvaluateTransaction("QueryRecords", selector)
	if err != nil {
		fmt.Printf("failed to evaluate transaction: %v\n", err)
		return ""
	}
	result := formatJSON(evaluateResult)
	return result
}

func UpdateReliabilityRecord(dataSourceID string, reliabilityScore float32, isDelta bool) {
	_, err := ClientContract.SubmitTransaction("UpdateReliabilityScore", dataSourceID, fmt.Sprintf("%f", reliabilityScore), fmt.Sprintf("%t", isDelta))
	if err != nil {
		fmt.Printf("failed to submit transaction: %v\n", err)
		return
	}
}

func GetHistoryForRecord(logID string) string {
	evaluateResult, err := ClientContract.EvaluateTransaction("GetHistoryForRecord", logID)
	if err != nil {
		fmt.Printf("failed to evaluate transaction: %v\n", err)
		return ""
	}
	result := formatJSON(evaluateResult)
	return result
}

// // main function
func main() {
	InitGateway()
	fmt.Println("InitGateway")
	// testInitLedger()
	// fmt.Println("testInitLedger")
	// testCreateLogRecord()
	// fmt.Println("testCreateLogRecord done")

	// selector := `{"selector": {"logID": "test_log_id"}}`
	// getRecordWithSelector(selector)

	// selector = `{"selector": {"logID": "default0", "type": "reliability"}}`
	// getRecordWithSelector(selector)

	// selector = `{"selector": {"logID": "default0", "type": "log"}}`
	// getRecordWithSelector(selector)

	GetAllReliabilityRecords()

	GetAllLogRecords()

	GetLogRecord("default0-reranker0")
	GetReliabilityRecord("default0")

	UpdateReliabilityRecord("default0", 0.9, true)
	GetReliabilityRecord("default0")
	GetHistoryForRecord("default0")

}
