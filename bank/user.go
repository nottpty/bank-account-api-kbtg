package bank

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type UserBankAct struct {
	User        User
	BankAccount []BankAccount
}
