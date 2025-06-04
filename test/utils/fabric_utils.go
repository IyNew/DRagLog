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
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()
}

// =====================================================================================================================
// For TESTING with asset-transfer-basic chaincode

// This type of transaction would typically only be run once by an application the first time it was started after its
// initial deployment. A new version of the chaincode deployed later would likely not need to run an "init" function.
func basicInitLedger() {
	fmt.Printf("\n--> Submit Transaction: InitLedger, function creates the initial set of log records on the ledger \n")

	_, err := ClientContract.SubmitTransaction("InitLedger")
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Printf("*** Transaction committed successfully\n")
}

// Evaluate a transaction to query ledger state.
func basicGetAllLogRecords() {
	fmt.Println("\n--> Evaluate Transaction: GetAllLogRecords, function returns all the current log records on the ledger")

	evaluateResult, err := ClientContract.EvaluateTransaction("GetAllLogRecords")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("*** Result:%s\n", result)
}

// Submit a transaction synchronously, blocking until it has been committed to the ledger.
func basicCreateLogRecord() {
	fmt.Printf("\n--> Submit Transaction: CreateLogRecord, creates new log record with logID, loggerID, input, inputFrom, output, outputTo, timestamp and reserved arguments \n")

	_, err := ClientContract.SubmitTransaction("CreateLogRecord", "test_log_id", "test_logger_id", "test_input", "test_input_from", "test_output", "test_output_to", "test_timestamp", "test_reserved")
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Printf("*** Transaction committed successfully\n")
}

func basicGetAllReliabilityRecords() {
	fmt.Println("\n--> Evaluate Transaction: GetAllReliabilityRecords, function returns all the current reliability records on the ledger")

	evaluateResult, err := ClientContract.EvaluateTransaction("GetAllReliabilityRecords")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("*** Result:%s\n", result)
}

func basicCreateReliabilityRecord() {
	fmt.Printf("\n--> Submit Transaction: CreateReliabilityRecord, creates new reliability record with datasourceID\n")

	_, err := ClientContract.SubmitTransaction("CreateReliabilityRecord", "test_data_source_id")
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Printf("*** Transaction committed successfully\n")
}

// END basic TESTING
// =====================================================================================================================

func createLogRecord(logID string, loggerID string, input string, inputFrom string, output string, outputTo string, timestamp string, reserved string) {
	// fmt.Printf("\n--> Submit Transaction: CreateRecord, creates new record with droneID, zip, flytime, flyrecord and reserved arguments \n")

	_, err := ClientContract.SubmitTransaction("CreateLogRecord", logID, loggerID, input, inputFrom, output, outputTo, timestamp, reserved)
	if err != nil {
		// panic(fmt.Errorf("failed to submit transaction: %w", err))
		fmt.Printf("failed to submit transaction: %v\n", err)
		// fmt.Printf("Please check if the record %s, %s, %s already exists\n", droneID, flytime, ConvertToRFC3339(flytime))
		return
	}

	// fmt.Printf("*** Transaction committed successfully\n")
}

func createLogRecordAsync(logID string, loggerID string, input string, inputFrom string, output string, outputTo string, timestamp string, reserved string) {
	// fmt.Printf("\n--> Submit Transaction: CreateRecord, creates new record with droneID, zip, flytime, flyrecord and reserved arguments \n")

	submitResult, commit, err := ClientContract.SubmitAsync("CreateLogRecord", client.WithArguments(logID, loggerID, input, inputFrom, output, outputTo, timestamp, reserved))
	if err != nil {
		// panic(fmt.Errorf("failed to submit transaction asynchronously: %w", err))
		fmt.Printf("failed to submit transaction asynchronously: %v\n", err)
		// fmt.Printf("Please check if the record %s, %s, %s already exists\n", logID, loggerID, input)
		return
	}

	fmt.Printf("\n*** Successfully submitted transaction to store the record: %s_%s. Info: %s\n", logID, timestamp, string(submitResult))
	// fmt.Println("*** Waiting for transaction commit.")

	if commitStatus, err := commit.Status(); err != nil {
		panic(fmt.Errorf("failed to get commit status: %w", err))
	} else if !commitStatus.Successful {
		panic(fmt.Errorf("transaction %s failed to commit with status: %d", commitStatus.TransactionID, int32(commitStatus.Code)))
	}

	// fmt.Printf("*** Transaction committed successfully\n")
}

func getAllLogRecords() string {
	// fmt.Println("\n--> Evaluate Transaction: GetAllRecords, function returns all the current records on the ledger")

	evaluateResult, err := ClientContract.EvaluateTransaction("GetAllLogRecords")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	// fmt.Printf("*** Result:%s\n", result)
	return result
}

func getAllReliabilityRecords() string {
	evaluateResult, err := ClientContract.EvaluateTransaction("GetAllReliabilityRecords")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	return result
}

func getAllReliabilityRecordsForOneDataSource(dataSourceID string) string {
	evaluateResult, err := ClientContract.EvaluateTransaction("GetAllReliabilityRecordsForOneDataSource", dataSourceID)
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	return result
}

func getLogRecord(logID string) string {
	evaluateResult, err := ClientContract.EvaluateTransaction("GetLogRecord", logID)
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	return result
}

func getRecordWithSelector(selector string) string {
	// fmt.Println("\n--> Evaluate Transaction: GetAllRecords, function returns all the current records on the ledger")

	evaluateResult, err := ClientContract.EvaluateTransaction("QueryRecords", selector)
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	// fmt.Printf("*** Result:%s\n", result)
	return result
}

func helloFabric() string {
	// fmt.Println("\n--> Evaluate Transaction: Hello, function returns a greeting message to check if the chaincode is alive")
	evaluateResult, err := ClientContract.EvaluateTransaction("Hello")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := string(evaluateResult)

	return result
}

// // main function
// func main() {
// 	InitGateway()
// 	basicInitLedger()
// 	basicCreateLogRecord()
// 	basicGetAllLogRecords()
// 	basicCreateReliabilityRecord()
// 	basicGetAllReliabilityRecords()
// 	basicGetAllReliabilityRecordsForOneDataSource("test_data_source_id")
// 	basicGetLogRecord("test_log_id")
// }
