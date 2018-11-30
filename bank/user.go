package bank

type User struct {
	ID        int
	FirstName string
	LastName  string
}

type UserBankAct struct {
	User        User
	BankAccount []BankAccount
}
