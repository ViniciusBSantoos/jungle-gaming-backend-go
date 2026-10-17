package domain

import "errors"

type Money struct {
	cents    int64
	currency string
}

func NewMoney(cents int64, currency string) (Money, error) {
	if cents < 0 {
		return Money{}, errors.New("valor negativo nao permitido")
	}

	if currency != "BRL" {
		return Money{}, errors.New("moeda nao suportada")
	}

	return Money{
		cents:    cents,
		currency: currency,
	}, nil
}

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, errors.New("moedas diferentes")
	}

	if other.cents > 0 && m.cents > (1<<63-1)-other.cents {
		return Money{}, errors.New("valor excede o limite")
	}

	return NewMoney(m.cents+other.cents, m.currency)
}

func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, errors.New("moedas diferentes")
	}

	if m.cents < other.cents {
		return Money{}, errors.New("saldo insuficiente")
	}

	return NewMoney(m.cents-other.cents, m.currency)
}
