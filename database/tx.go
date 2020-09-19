package database

// Account is an alias for customer represented in DB
type Account string

// NewAccount return new customer account
func NewAccount(value string) Account {
	return Account(value)
}

// Tx represent each transaction in database
type Tx struct {
	From  Account `json:"from"`
	To    Account `json:"to"`
	Value uint    `json:"value"`
	Data  string  `json:"data"`
}

// NewTx return new transaction
func NewTx(from Account, to Account, value uint, data string) Tx {
	return Tx{from, to, value, data}
}

// IsReward check if transaction is eligible for a reward
func (t Tx) IsReward() bool {
	return t.Data == "reward"
}
