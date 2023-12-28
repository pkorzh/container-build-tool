package main

import (
	"github.com/pkorzh/container-build-tool/pkg/storage"
)

func getStore() (storage.Store, error) {
	defaultStoreOptions := storage.DefaultStoreOptions()

	store, err := storage.New(defaultStoreOptions)

	if err != nil {
		return nil, err
	}

	return store, nil
}
