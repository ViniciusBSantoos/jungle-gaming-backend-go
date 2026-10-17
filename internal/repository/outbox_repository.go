package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func InsertOutboxEvent(
	ctx context.Context,
	tx pgx.Tx,
	id string,
	aggregateID string,
	eventType string,
	payload any,
) error {
	if eventType == "" {
		return fmt.Errorf("tipo de evento obrigatorio")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao gerar JSON: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO outbox_events (
            id,
            aggregate_id,
            event_type,
            payload
        ) VALUES ($1, $2, $3, $4::jsonb)`,
		id,
		aggregateID,
		eventType,
		string(data),
	)

	if err != nil {
		return fmt.Errorf("erro ao registrar outbox: %w", err)
	}

	return nil
}
