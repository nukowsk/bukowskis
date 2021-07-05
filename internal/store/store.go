package store

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/mitchellh/hashstructure/v2"
)

func storeID(tx *types.Transaction) (string, error) {
	objectHash, err := hashstructure.Hash(tx.Hash().Hex(), hashstructure.FormatV2, nil)
	if err != nil {
		return "", fmt.Errorf("Failed to hash raw transaction: %s\n", err)
	}
	return strconv.FormatUint(objectHash, 10), nil
}

// Hash and ID are confusing and should be given more distinctive names
type LogEntry struct {
	Hash        string
	Transaction string
	Auction     string
	Timestamp   time.Time
}

func NewLogEntry(tx *types.Transaction) (LogEntry, error) {
	hash, err := storeID(tx)
	if err != nil {
		return LogEntry{}, err
	}

	return LogEntry{
		Hash:        hash,
		Transaction: tx.Hash().Hex(),
		Auction:     "open",
		Timestamp:   time.Now(), // XXX: Probably want to pass this in
	}, nil
}

type Store interface {
	Save(*LogEntry) error
	Query(time.Time, time.Time) ([]LogEntry, error)
	Close()
}

type Firestore struct {
	client *firestore.Client
}

func NewFirestore(projectId string) (*Firestore, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("Fatal firebase error :%s", err)
	}

	return &Firestore{
		client: client,
	}, nil
}

func (f *Firestore) Save(logEntry *LogEntry) error {
	ctx := context.Background()
	collection := f.client.Collection("txs").Doc(logEntry.Hash)
	_, err := collection.Create(ctx, logEntry)
	if err != nil {
		return fmt.Errorf("Failed to add transaction: %v", err)
	}

	return nil
}

func (f *Firestore) Query(from time.Time, to time.Time) ([]LogEntry, error) {
	// TODO
	return []LogEntry{}, nil
}

func (f *Firestore) Close() {
	f.client.Close()
}

type Local struct {
	items []LogEntry
}

func NewLocal() (*Local, error) {
	return &Local{
		items: []LogEntry{},
	}, nil
}

func (l *Local) Save(logEntry *LogEntry) error {
	// TODO
	return nil
}

func (l *Local) Query(from time.Time, to time.Time) ([]LogEntry, error) {
	// TODO
	return []LogEntry{}, nil
}

func (l *Local) Close() {
	return
}
