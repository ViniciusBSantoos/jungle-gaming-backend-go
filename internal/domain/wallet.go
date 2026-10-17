package domain

import (
	"errors"
	"time"
)

type Wallet struct {
	ID        string
	PlayerID  string
	Balance   Money
	Version   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewWallet(id, playerID string, balance Money) (Wallet, error) {
	if id == "" || playerID == "" {
		return Wallet{}, errors.New("identificacao obrigatoria")
	}

	if balance.currency == "" {
		return Wallet{}, errors.New("saldo invalido")
	}

	if balance.cents < 0 {
		return Wallet{}, errors.New("saldo negativo")
	}

	now := time.Now().UTC()

	return Wallet{
		ID:        id,
		PlayerID:  playerID,
		Balance:   balance,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (w *Wallet) Credit(amount Money) error {
	if amount.cents <= 0 {
		return errors.New("credito deve ser positivo")
	}

	newBalance, err := w.Balance.Add(amount)
	if err != nil {
		return err
	}

	w.Balance = newBalance
	w.Version++
	w.UpdatedAt = time.Now().UTC()

	return nil
}

func (w *Wallet) Debit(amount Money) error {
	if amount.cents <= 0 {
		return errors.New("debito deve ser positivo")
	}

	newBalance, err := w.Balance.Subtract(amount)
	if err != nil {
		return err
	}

	w.Balance = newBalance
	w.Version++
	w.UpdatedAt = time.Now().UTC()

	return nil
}
