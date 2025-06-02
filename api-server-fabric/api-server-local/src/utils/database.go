package utils

import (
	"context"
	"fmt"
	"log"

	kivik "github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/couchdb" // The CouchDB driver
)

// PutStateToDB stores the state in the database
func PutStateToDB(db *kivik.DB, stateDBO *StateDBObject) error {
	// Create a new document in the database
	_, err := db.Put(context.Background(), stateDBO.StateID, stateDBO)
	if err != nil {
		return fmt.Errorf("failed to store state: %w", err)
	}
	return nil
}

// GetStateWithStateID retrieves the state from the database
func GetStateWithStateID(db *kivik.DB, stateID string) (*StateDBObject, error) {
	// Get the document from the database
	var stateDBO StateDBObject
	err := db.Get(context.Background(), stateID).ScanDoc(&stateDBO)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve state: %w", err)
	}
	return &stateDBO, nil
}

// GetStatesWithSelector retrieves the states from the database
func GetStatesWithSelector(db *kivik.DB, rawSelector map[string]interface{}) ([]StateDBObject, error) {
	selector := map[string]interface{}{
		"selector": rawSelector,
	}
	log.Printf("Selector: %v", selector)
	rows := db.Find(context.Background(), selector)
	defer rows.Close()

	var stateDBO []StateDBObject
	for rows.Next() {
		var state StateDBObject
		if err := rows.ScanDoc(&state); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		stateDBO = append(stateDBO, state)
	}
	return stateDBO, nil
}

func UpdateState(db *kivik.DB, sourceID string, creditChange float64) error {
	// Get the current state
	state, err := GetStateWithStateID(db, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}

	state.Credits += creditChange

	err = PutStateToDB(db, state)
	if err != nil {
		return fmt.Errorf("failed to put state: %w", err)
	}

	return nil
}
