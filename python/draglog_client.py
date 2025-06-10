import requests
from typing import List, Optional, Dict, Any
from dataclasses import dataclass
from datetime import datetime
from ctypes import c_float as float32

@dataclass
class LogRecord:
    logID: str  # logID or dataSourceID
    loggerID: str
    type: str  # log or reliability
    input: str
    inputFrom: str
    output: str
    outputTo: str
    reliabilityScore: float  # -1 for log
    timestamp: str
    reserved: str

@dataclass
class LogRecordInput:
    logID: str  # logID or dataSourceID
    loggerID: str
    input: str
    inputFrom: str
    output: str
    outputTo: str
    timestamp: str
    reserved: str
    type: str = "log"  
    reliabilityScore: float32 = -1

@dataclass
class LogRecordHistory:
    record: Optional[LogRecord]
    timestamp: str
    txID: str
    isDelete: bool

class DragLogClient:
    def __init__(self, base_url: str = "http://localhost:8080"):
        """Initialize the DragLog client.
        
        Args:
            base_url: Base URL of the DragLog API server
        """
        self.base_url = base_url.rstrip('/')
        
    def _make_request(self, method: str, endpoint: str, **kwargs) -> Dict[str, Any]:
        """Make an HTTP request to the API server.
        
        Args:
            method: HTTP method (GET, POST, PUT)
            endpoint: API endpoint
            **kwargs: Additional arguments for requests
            
        Returns:
            Response data as dictionary
        """
        url = f"{self.base_url}/{endpoint.lstrip('/')}"
        response = requests.request(method, url, **kwargs)
        response.raise_for_status()
        
        # Return empty dict if response is empty
        if not response.content:
            return {}
            
        return response.json()
    
    def init_ledger(self) -> None:
        """Initialize the ledger."""
        self._make_request('GET', '/init-ledger')
    
    def create_log_record(self, record: LogRecordInput) -> None:
        """Create a new log record.
        
        Args:
            record: LogRecordInput object containing the record details
        """
        # print({"body": record.__dict__})
        self._make_request('POST', '/create-log-record', json=record.__dict__)
    
    def create_reliability_record(self, data_source_id: str, digest: str) -> None:
        """Create a new reliability record.
        
        Args:
            data_source_id: ID of the data source
            digest: Digest value
        """
        self._make_request('POST', '/create-reliability-record', 
                          json={"dataSourceID": data_source_id, "digest": digest})
    
    def get_all_log_records(self) -> List[LogRecord]:
        """Get all log records.
        
        Returns:
            List of LogRecord objects
        """
        response = self._make_request('GET', '/get-all-log-records')

        return [LogRecord(**record) for record in response['records']]
    
    def get_all_reliability_records(self) -> List[LogRecord]:
        """Get all reliability records.
        
        Returns:
            List of LogRecord objects
        """
        response = self._make_request('GET', '/get-all-reliability-records')
        return [LogRecord(**record) for record in response['records']]
    
    def get_log_record(self, log_id: str) -> LogRecord:
        """Get a specific log record.
        
        Args:
            log_id: ID of the log record
            
        Returns:
            LogRecord object
        """
        response = self._make_request('GET', f'/get-log-record/{log_id}')
        return LogRecord(**response['records'][0])
    
    def get_reliability_record(self, data_source_id: str) -> LogRecord:
        """Get a specific reliability record.
        
        Args:
            data_source_id: ID of the data source
            
        Returns:
            LogRecord object
        """
        response = self._make_request('GET', f'/get-reliability-record/{data_source_id}')
        return LogRecord(**response['records'][0])
    
    def update_reliability_record(self, data_source_id: str, reliability_score: float) -> None:
        """Update a reliability record's score.
        
        Args:
            data_source_id: ID of the data source
            reliability_score: New reliability score
        """
        self._make_request('PUT', f'/update-reliability-record/{data_source_id}',
                          json={"reliabilityScore": reliability_score})
    
    def get_history_for_record(self, log_id: str) -> List[LogRecordHistory]:
        """Get the history of a record.
        
        Args:
            log_id: ID of the log record
            
        Returns:
            List of LogRecordHistory objects
        """
        response = self._make_request('GET', f'/get-history-for-record/{log_id}')
        return [LogRecordHistory(
            record=LogRecord(**history['record']) if history['record'] else None,
            timestamp=history['timestamp'],
            txID=history['txID'],
            isDelete=history['isDelete']
        ) for history in response['history']]

# Example usage:
if __name__ == "__main__":
    # Create client with custom server address
    client = DragLogClient("http://localhost:8080")
    
    # Initialize ledger
    client.init_ledger()
    
    # Create a log record
    record = LogRecord(
        logID="test123",
        loggerID="logger1",
        type="log",
        input="test input",
        inputFrom="source1",
        output="test output",
        outputTo="destination1",
        reliabilityScore=-1.0,
        timestamp=datetime.now().isoformat(),
        reserved=""
    )
    client.create_log_record(record)
    
    # Get all log records
    records = client.get_all_log_records()
    for record in records:
        print(f"Found record: {record.logID}")