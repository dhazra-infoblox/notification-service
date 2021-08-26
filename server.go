package main

import (
	"fmt"
	"io"
	"net/http"
)

type ConfigStore interface {
	Fetch(logtype LoggerType) *Config
	Update(io.ReadCloser) *Config // Creates if not already present

	// ToDo: Delete
}

// Out Of Scope from LLD perspective
// API Handler for Config CRUD operations
func ConfigServer(store ConfigStore) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			fmt.Fprintf(rw, store.Fetch(INFO).String())
		case http.MethodPut, http.MethodPost:
			fmt.Fprintf(rw, store.Update(r.Body).String())
		}
	}
}

type StubConfigStore struct {
	configs []*Config
}

func NewStubConfigStore() *StubConfigStore {
	var configs []*Config
	configs = append(configs, NewConfig(INFO, 100, 100, 100), NewConfig(CRITICAL, 10, 100, 100), NewConfig(WARNING, 20, 100, 100))
	return &StubConfigStore{configs}
}

func (s *StubConfigStore) Fetch(logtype LoggerType) *Config {
	for _, v := range s.configs {
		if v.LogType == logtype {
			return v
		}
	}
	return nil
}

func (s *StubConfigStore) Update(_ io.ReadCloser) *Config {
	panic("not implemented") // TODO: Implement
}

// ToDo: Similar to above, Create API Handlers for Topic and LogStream
