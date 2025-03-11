package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Secret struct {
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"-"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Key        string    `json:"key"`
	Value      string    `json:"value"`
	Passphrase []byte    `json:"-"`
}

type SecretModel struct {
	DB            *sql.DB
	encryptionKey string
}

func (m SecretModel) Insert(secret *Secret) error {
	query := `
	INSERT INTO secrets (expires_at, key, value, passphrase)
	VALUES (
		$1, 
		$2, 
		pgp_sym_encrypt($3,$5), 
		pgp_sym_encrypt($4,$5)
		)
	RETURNING id, created_at
	`

	args := []any{secret.ExpiresAt, secret.Key, secret.Value, secret.Passphrase, m.encryptionKey}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&secret.ID, &secret.CreatedAt)
}

func (m SecretModel) Get(key string) (*Secret, error) {
	if key == "" {
		return nil, ErrRecordNotFound
	}

	query := `
	SELECT id, created_at, expires_at, key, pgp_sym_decrypt(value::BYTEA, $2), pgp_sym_decrypt(passphrase, $2)
	FROM secrets
	WHERE key = $1`

	var secret Secret
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, key, m.encryptionKey).Scan(
		&secret.ID,
		&secret.CreatedAt,
		&secret.ExpiresAt,
		&secret.Key,
		&secret.Value,
		&secret.Passphrase,
	)
	fmt.Println(secret.Passphrase)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &secret, nil
}

func (m SecretModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
	DELETE FROM secrets
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m SecretModel) CleanupExpired() error {
	query := `
	DELETE FROM secrets
	WHERE expires_at <= NOW()
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}
