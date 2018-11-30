package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"bank-account-api-kbtg/bank"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func connectDB() {
	var err error
	const connStr = "admin:@tcp(127.0.0.1:3306)/bank_account_api?parseTime=true"
	db, err = sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *UserServiceImp) All() ([]bank.User, error) {
	rows, err := s.db.Query("SELECT * FROM user")
	if err != nil {
		return nil, err
	}
	users := []bank.User{} // set empty slice without nil
	for rows.Next() {
		var user bank.User
		err := rows.Scan(&user.ID, &user.FirstName, &user.LastName)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *UserServiceImp) Insert(user *bank.User) (int64, error) {
	res, err := s.db.Exec("INSERT INTO user (first_name, last_name) values (?, ?)", user.FirstName, user.LastName)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

func (s *UserServiceImp) InsertAccount(bankAccount *bank.BankAccount) (int64, error) {
	res, err := s.db.Exec("INSERT INTO bank_account (user_id, account_number, name, balance) values (?, ?, ?, ?)", bankAccount.UserID, bankAccount.AccountNumber, bankAccount.Name, bankAccount.Balance)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

func (s *UserServiceImp) GetByID(id int) (*bank.User, error) {
	stmt := "SELECT * FROM user WHERE id = ?"
	row := s.db.QueryRow(stmt, id)
	var user bank.User
	err := row.Scan(&user.ID, &user.FirstName, &user.LastName)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserServiceImp) GetAccountByID(id int) ([]bank.BankAccount, error) {
	rows, err := s.db.Query("SELECT * FROM bank_account WHERE user_id = ?", id)
	if err != nil {
		return nil, err
	}
	accounts := []bank.BankAccount{} // set empty slice without nil
	for rows.Next() {
		var account bank.BankAccount
		err := rows.Scan(&account.ID, &account.UserID, &account.AccountNumber, &account.Name, &account.Balance)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (s *UserServiceImp) DeleteByID(id int) error {
	stmt := "DELETE FROM user WHERE id = ?"
	_, err := s.db.Exec(stmt, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserServiceImp) Update(id int, firstName string, lastName string) (*bank.User, error) {
	stmt := "UPDATE user SET first_name = ?, last_name = ? WHERE id = ?"
	_, err := s.db.Exec(stmt, firstName, lastName, id)
	if err != nil {
		return nil, err
	}
	return s.GetByID(id)
}

type Server struct {
	db          *sql.DB
	userService UserService
}

func (s *Server) All(c *gin.Context) {
	todos, err := s.userService.All()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"object":  "error",
			"message": fmt.Sprintf("db: query error: %s", err),
		})
		return
	}
	c.JSON(http.StatusOK, todos)
}

func (s *Server) Create(c *gin.Context) {
	var user bank.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"object":  "error",
			"message": fmt.Sprintf("json: wrong params: %s", err),
		})
		return
	}

	id, err := s.userService.Insert(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	user.ID = int(id)

	c.JSON(http.StatusCreated, user)
}

func (s *Server) CreateAccount(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("id"))
	user, err := s.userService.GetByID(userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	bankAccount := bank.BankAccount{
		UserID: user.ID,
	}
	err = c.ShouldBindJSON(&bankAccount)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"object":  "error",
			"message": fmt.Sprintf("json: wrong params: %s", err),
		})
		return
	}

	id, err := s.userService.InsertAccount(&bankAccount)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	bankAccount.ID = int(id)

	c.JSON(http.StatusCreated, bankAccount)
}

func (s *Server) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, err := s.userService.GetByID(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (s *Server) GetAccountByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	accounts, err := s.userService.GetAccountByID(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"object":  "error",
			"message": fmt.Sprintf("db: query error: %s", err),
		})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

func (s *Server) DeleteByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := s.userService.DeleteByID(id); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
}

func (s *Server) Update(c *gin.Context) {
	h := map[string]string{}
	if err := c.ShouldBindJSON(&h); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	user, err := s.userService.Update(id, h["first_name"], h["last_name"])
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

type UserService interface {
	All() ([]bank.User, error)
	Insert(user *bank.User) (int64, error)
	GetByID(id int) (*bank.User, error)
	DeleteByID(id int) error
	Update(id int, firstName string, lastName string) (*bank.User, error)
	InsertAccount(bankAccount *bank.BankAccount) (int64, error)
	GetAccountByID(id int) ([]bank.BankAccount, error)
}

type UserServiceImp struct {
	db *sql.DB
}

func setupRoute(s *Server) *gin.Engine {
	r := gin.Default()
	users := r.Group("/users")
	// admin := r.Group("/bankaccounts")

	// admin.Use(gin.BasicAuth(gin.Accounts{
	// 	"admin": "1234",
	// }))
	// todos.Use(s.AuthTodo)
	users.GET("/", s.All)
	users.POST("/", s.Create)

	users.GET("/:id", s.GetByID)
	users.PUT("/:id", s.Update)
	users.DELETE("/:id", s.DeleteByID)

	users.POST("/:id/bankAccounts", s.CreateAccount)
	users.GET("/:id/bankAccounts", s.GetAccountByID)
	return r
}

func StartServer() {
	// db, err := sql.Open("mysql", os.Getenv("DATABASE_URL"))
	connectDB()

	s := &Server{
		db: db,
		userService: &UserServiceImp{
			db: db,
		},
	}

	r := setupRoute(s)

	// r.Run(":" + os.Getenv("PORT"))
	r.Run(":8000")
}
