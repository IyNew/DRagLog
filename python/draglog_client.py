import requests
from typing import List, Optional, Dict, Any
from dataclasses import dataclass
from datetime import datetime
from ctypes import c_float as float32
import json
import os
import hashlib


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
    def __init__(self, base_url: str = "http://localhost:8080", debug: bool = False, debug_file: str = "draglog_debug.json", local: bool = False):
        """Initialize the DragLog client.
        
        Args:
            base_url: Base URL of the DragLog API server
            debug: Whether to enable debug mode and store records in a file
            debug_file: Path to the debug log file
            local: Whether to store records locally without making API requests
        """
        self.base_url = base_url.rstrip('/')
        self.debug = debug
        self.debug_file = debug_file
        self.local = local
        if self.debug or self.local:
            # Initialize debug file if it doesn't exist
            if not os.path.exists(self.debug_file):
                with open(self.debug_file, 'w') as f:
                    json.dump([], f)
        
    def _make_request(self, method: str, endpoint: str, **kwargs) -> Dict[str, Any]:
        """Make an HTTP request to the API server.
        
        Args:
            method: HTTP method (GET, POST, PUT)
            endpoint: API endpoint
            **kwargs: Additional arguments for requests
            
        Returns:
            Response data as dictionary or empty dict if request fails
        """
        if self.local:
            # In local mode, return empty dict for GET requests
            if method == 'GET':
                return {'records': []}
            # For other methods, just write to debug file and return empty dict
            if 'json' in kwargs:
                self._write_to_debug_file(kwargs['json'], f'{method.lower()}_{endpoint.replace("/", "_")}')
            return {}
            
        try:
            url = f"{self.base_url}/{endpoint.lstrip('/')}"
            response = requests.request(method, url, **kwargs)
            response.raise_for_status()
            
            # Return empty dict if response is empty
            if not response.content:
                return {}
                
            return response.json()
        except requests.RequestException as e:
            print(f"Warning: Failed to connect to server at {self.base_url}: {str(e)}")
            return {}
    
    def init_ledger(self) -> None:
        """Initialize the ledger."""
        self._make_request('GET', '/init-ledger')
    
    def _write_to_debug_file(self, record: Dict[str, Any], operation: str) -> None:
        """Write a record to the debug file in JSONL format.
        
        Args:
            record: The record to write
            operation: The operation performed (create, update, etc.)
        """
        if not (self.debug or self.local):
            return
            
        try:
            debug_record = {
                'timestamp': datetime.now().isoformat(),
                'operation': operation,
                'record': record
            }
            
            # Append the record as a single line
            with open(self.debug_file, 'a') as f:
                f.write(json.dumps(debug_record) + '\n')
        except Exception as e:
            print(f"Error writing to debug file: {e}")

    def _read_debug_records(self) -> List[Dict[str, Any]]:
        """Read all records from the debug file.
        
        Returns:
            List of records from the debug file
        """
        if not os.path.exists(self.debug_file):
            return []
            
        records = []
        try:
            with open(self.debug_file, 'r') as f:
                for line in f:
                    if line.strip():  # Skip empty lines
                        records.append(json.loads(line))
        except Exception as e:
            print(f"Error reading debug file: {e}")
        return records

    def create_log_record(self, record: LogRecordInput) -> None:
        """Create a new log record.
        
        Args:
            record: LogRecordInput object containing the record details
        """
        record.type = "log"
        record_dict = record.__dict__
        self._make_request('POST', '/create-log-record', json=record_dict)
        if self.debug:
            self._write_to_debug_file(record_dict, 'create_log_record')
    
    def create_feedback_record(self, record: LogRecordInput) -> None:
        """Create a new feedback record.
        
        Args:
            record: LogRecordInput object containing the record details
        """
        record.type = "feedback"
        record_dict = record.__dict__
        self._make_request('POST', '/create-feedback-record', json=record_dict)
        if self.debug:
            self._write_to_debug_file(record_dict, 'create_feedback_record')
    
    def create_reliability_record(self, data_source_id: str, digest: str, reserved: str) -> None:
        """Create a new reliability record.
        
        Args:
            data_source_id: ID of the data source
            digest: Digest value
            reserved: Reserved value
        """
        # check if the record already exists
        # if self.get_reliability_record(data_source_id):
        #     print(f"Reliability record already exists for {data_source_id}")
        #     return
        record_dict = {"dataSourceID": data_source_id, "digest": digest, "reserved": reserved}
        self._make_request('POST', '/create-reliability-record', json=record_dict)
        if self.debug:
            # print(f"Creating reliability record: {record_dict} in debug mode")
            self._write_to_debug_file(record_dict, 'create_reliability_record')

    def create_reliability_record_async(self, data_source_id: str, digest: str, reserved: str) -> None:
        """Create a new reliability record asynchronously.
        
        Args:
            data_source_id: ID of the data source
            digest: Digest value
            reserved: Reserved value
        """
        record_dict = {"dataSourceID": data_source_id, "digest": digest, "reserved": reserved}
        self._make_request('POST', '/create-reliability-record-async', json=record_dict)
        if self.debug:
            self._write_to_debug_file(record_dict, 'create_reliability_record_async')
    
    def get_all_log_records(self) -> List[LogRecord]:
        """Get all log records.
        
        Returns:
            List of LogRecord objects
        """
        if self.local:
            records = self._read_debug_records()
            return [LogRecord(**record['record']) for record in records 
                   if record['operation'].startswith('create_log_record')]
            
        response = self._make_request('GET', '/get-all-log-records')
        return [LogRecord(**record) for record in response['records']]
    
    def get_all_reliability_records(self) -> List[LogRecord]:
        """Get all reliability records.
        
        Returns:
            List of LogRecord objects
        """
        if self.local:
            records = self._read_debug_records()
            return [LogRecord(**record['record']) for record in records 
                   if record['operation'].startswith('create_reliability_record')]
            
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
    
    def update_reliability_record(self, data_source_id: str, reliability_score: float, is_delta: bool) -> None:
        """Update a reliability record's score.
        
        Args:
            data_source_id: ID of the data source
            reliability_score: New reliability score
        """
        # check if the record already exists
        # if not self.get_reliability_record(data_source_id):
        #     print(f"Reliability record does not exist for {data_source_id}")
        #     return
        record_dict = {"reliabilityScore": reliability_score, "isDelta": is_delta}
        # print("update reliability record", record_dict)
        self._make_request('PUT', f'/update-reliability-record/{data_source_id}', json=record_dict)
        record_dict["dataSourceID"] = data_source_id
        if self.debug or self.local:
            self._write_to_debug_file(record_dict, 'update_reliability_record')
    
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


def doc_to_sha256(doc):
    """
    Transform a document of any length to a SHA256 hash.
    
    Args:
        doc (str): The document text to hash
        
    Returns:
        str: SHA256 hash as a hexadecimal string
    """
    # Encode the document to bytes (SHA256 requires bytes input)
    doc_bytes = doc.encode('utf-8')
    
    # Create SHA256 hash object
    sha256_hash = hashlib.sha256()
    
    # Update the hash object with the document bytes
    sha256_hash.update(doc_bytes)
    
    # Return the hexadecimal representation of the hash
    return sha256_hash.hexdigest()