# DRagLog
A basic logging system for drag



# System Workflow
```mermaid
sequenceDiagram
    participant BC as Blockchain
    participant DB as DataSource (with retriever) #99FF99
    participant DB2 as DataSource2 (with retriever) #99FF99
    participant DB3 as DataSource3 (with retriever) #99FF99
    participant RR as Re-ranker
    participant LLM
    Participant User

    User ->> LLM: Question
    Note over User, DB: Topic query omitted for now

    BC ->> DB: Retrieve reliability score from the blockchain query
    BC ->> DB2: Retrieve reliability score from the blockchain query
    BC ->> DB3: Retrieve reliability score from the blockchain query

    DB ->> RR: Retrieved documents
    DB -->> BC: Datasource ouput log
    DB2 ->> RR: Retrieved documents
    DB2 -->> BC: Datasource ouput log
    DB3 ->> RR: Retrieved documents
    DB3 -->> BC: Datasource ouput log
    
    RR ->> LLM: Re-ranked content
    RR ->> BC: Re-ranker output log

    LLM ->> User: Answer
    User ->> BC: Feedback with evaluation

```

# Data structures 
Reliability Score 
```json
{
    "SourceID": The identifier for the data source,
    "Score": The score for the data source,
    "TimeStamp": The timestamp of the last update
}
```

Data Source Output Log: assuming only one document provided by each Datasource
```json
{
    "Source_ID":  The identifier for the data source,
    "Digest": The digest of the output,
    "TimeStamp": The timestamp of the log
}
```

<!-- Omit for document-level traceback -->
<!-- Re-ranker Output Log
```json
{
    "Input_Source_IDs": The list of input source IDs,
    "Input_Digests": The list of input digests, 
    "Output_Source_IDs": The list of output,
    "Output_Digests": The list of output digests
}
``` -->

LLM Output Log
```json
{
    "LLM_ID": LLM identifier,
    "Input_Digests": The list of input digests,
    "Output_Digest": The digest of output
}
```

On-chain state definition
```json
{
    "State_ID": Datasoruce identifier or composite identifier for logs,
    "Type": Type of the state info,
    "Content": Score/Digest,
    "TimeStamp": The timestamp for the last update,
    "Reserved": Reserved space for other use
}
```

# Todo


