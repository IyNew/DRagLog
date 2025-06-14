REPO_SRC = $(shell pwd)
FABRIC_TEST_NETWORK_SRC = $(REPO_SRC)/fabric-samples/test-network
CONTRACT_SRC = $(REPO_SRC)/log-storage/chaincode-go
API_SERVER_SRC = $(REPO_SRC)/api-server
# check-prerequisite:
# 	@echo "Check prerequisite"
# 	@chmod +x pre-requisite.sh
# 	@./pre-requisite.sh


# Download a script to the current path
download_script:
	curl -sSLO https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh && chmod +x install-fabric.sh

install: download_script
	@echo "Installing Fabric"
	./install-fabric.sh d s b
	@echo "Installation complete"

network_up:
	cd $(FABRIC_TEST_NETWORK_SRC) && ./network.sh up createChannel

network_up_couchdb:
	cd $(FABRIC_TEST_NETWORK_SRC) && ./network.sh up createChannel -s couchdb

# Ignore the couchdb setting for now, check the performance later
draglog_deploy: down network_up
	cd $(FABRIC_TEST_NETWORK_SRC) && ./network.sh deployCC -ccn basic -ccp $(CONTRACT_SRC) -ccl go

draglog_couchdb_deploy: down network_up_couchdb
	cd $(FABRIC_TEST_NETWORK_SRC) && ./network.sh deployCC -ccn basic -ccp $(CONTRACT_SRC) -ccl go 

draglog_contract_update:
	cd $(FABRIC_TEST_NETWORK_SRC) && ./network.sh deployCC -ccn basic -ccp $(CONTRACT_SRC) -ccl go 

api_server: 
	cd $(API_SERVER_SRC) && nohup go run main.go > api_server.log 2>&1 &

all: draglog_couchdb_deploy api_server

down:
	cd $(FABRIC_TEST_NETWORK_SRC) && ./network.sh down
	-kill -9 $(lsof -t -i:8080)



# Stop the API server
stop_api_server:
	@echo "Stopping API server"
	-kill -9 $(lsof -t -i:8080)

# Clean command to remove all materials
clean:
	rm -rf fabric-samples
	rm -f install-fabric.sh