package service

type BetResult struct {
	TransactionID string
	Status        string
	BalanceCents  int64
	Replayed      bool
}
