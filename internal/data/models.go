package data

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Secrets SecretModel
}

func NewModels(db *sql.DB, encryptionKey string) Models {
	secretModel := SecretModel{db, encryptionKey}

	// Cleaning up expired secrets each minute
	go func(secretModel *SecretModel) {
		for {
			time.Sleep(1 * time.Minute)
			secretModel.CleanupExpired()
		}
	}(&secretModel)

	return Models{
		Secrets: secretModel,
	}
}
