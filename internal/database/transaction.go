package database

import (
	"context"
	"database/sql"
	"fmt"
)

type Database interface {
	DBTX

	WithinTransaction(ctx context.Context, fn func(tx DBTX) error) error
}

type Client struct {
	*sql.DB
}

func NewClient(db *sql.DB) *Client {
	return &Client{DB: db}
}

func (c *Client) WithinTransaction(ctx context.Context, fn func(tx DBTX) error) error {
	tx, err := c.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
