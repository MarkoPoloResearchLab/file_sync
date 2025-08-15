package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type syncState struct {
	FileEntry map[string]stateEntry `json:"file_entry"`
}

type stateEntry struct {
	AncestorHex string `json:"ancestor_hex"`
}

type stateStore struct {
	StatePath string
	AncDir    string
}

func createOrOpenStateStore(stateDir string) (*stateStore, *syncState, error) {
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return nil, nil, err
	}
	ancDir := filepath.Join(stateDir, "ancestors")
	if err := os.MkdirAll(ancDir, 0o755); err != nil {
		return nil, nil, err
	}
	statePath := filepath.Join(stateDir, "state.json")
	state := &syncState{FileEntry: map[string]stateEntry{}}

	if _, err := os.Stat(statePath); errors.Is(err, fs.ErrNotExist) {
		if err := os.WriteFile(statePath, []byte(`{"file_entry":{}}`), 0o644); err != nil {
			return nil, nil, err
		}
	} else if err == nil {
		data, readErr := os.ReadFile(statePath)
		if readErr != nil {
			return nil, nil, readErr
		}
		if unmarshalErr := json.Unmarshal(data, state); unmarshalErr != nil {
			return nil, nil, unmarshalErr
		}
	} else {
		return nil, nil, err
	}

	return &stateStore{StatePath: statePath, AncDir: ancDir}, state, nil
}

func (s *stateStore) save(state *syncState) error {
	tmpPath := s.StatePath + ".tmp"
	data, marshalErr := json.MarshalIndent(state, "", "  ")
	if marshalErr != nil {
		return marshalErr
	}
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, s.StatePath)
}

func digestBytes(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func (s *stateStore) ancestorBytes(hexDigest string) ([]byte, error) {
	path := filepath.Join(s.AncDir, hexDigest)
	return os.ReadFile(path)
}

func (s *stateStore) ensureAncestorStored(content []byte) (string, error) {
	hexDigest := digestBytes(content)
	path := filepath.Join(s.AncDir, hexDigest)
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		if writeErr := os.WriteFile(path, content, 0o644); writeErr != nil {
			return "", writeErr
		}
	} else if err != nil {
		return "", err
	}
	return hexDigest, nil
}
