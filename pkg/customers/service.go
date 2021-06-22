package customers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"

)

var ErrNotFound = errors.New("item not found")

var ErrInternal = errors.New("internal error")
var ErrNoSuchUser = errors.New("no such user")
var ErrPhoneUsed = errors.New("phone already registered")
var ErrInvalidPassword = errors.New("invalid password")
var ErrTokenNotFound = errors.New("token not found")
var ErrTokenExpired = errors.New("token expired")

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Customer struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	Phone    string    `json:"phone"`
	Password string    `json:"password"`
	Active   bool      `json:"active"`
	Created  time.Time `json:"created"`
}

type Registration struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func (s *Service) Register(ctx context.Context, registration *Registration) (*Customer, error) {
	var err error
	item := &Customer{}

	hash, err := bcrypt.GenerateFromPassword([]byte(registration.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrInternal
	}
	log.Print(hash)

	err = s.pool.QueryRow(ctx, `
	INSERT INTO customers (name, phone, password)
	VALUES ($1,$2,$3)
	ON CONFLICT (phone) DO NOTHING RETURNING id,name,phone, active, created
	`, registration.Name, registration.Phone, hash).Scan(
		&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrInternal
	}
	if err != nil {
		return nil, ErrInternal
	}

	return item, nil
}

type Auth struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}

type Sales struct {
	 ID 	       int64  		`json:"id"`
	 Name          string 		`json:"name"`
	 Price         int    		`json:"price"`
	 Qty           int 	  		`json:"qty"`
	 Created       time.Time 	`json:"created"`
}



type Product struct{
	ID 		int64	`json:"id"`
	Name 	string  `json:"name"`
	Price   int 	`json:"price"`
	Qty		int 	`json:"qty"`
}

func (s *Service) Products(ctx context.Context) ([]*Product, error) {
	items :=make([]*Product, 0)

	rows, err := s.pool.Query(ctx, `
	SELECT id,name, price,qty FROM products WHERE active ORDER BY id LIMIT 500
	`)
	if errors.Is(err, pgx.ErrNoRows) {
		return items, nil 
	}
	if err != nil {
		return nil, ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		item := &Product{}
		err = rows.Scan(&item.ID, &item.Name, &item.Price, &item.Qty)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil,err
	}
	return items, nil 
} 

//find Id customers via Token
func (s *Service) IDByToken(ctx context.Context, token string) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx,`
	SELECT customer_id FROM customers_tokens WHERE token =$1
	`, token).Scan(&id)

	if err == pgx.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, ErrInternal
	}
	return id,nil 
}

func (s *Service) Purchases(ctx context.Context, id int64) ([]*Sales, error) {
	sales :=make([]*Sales, 0)

	rows, err := s.pool.Query(ctx, `
	SELECT sp.id, sp.name, sp.price,sp.qty,sp.created 
	FROM sale_positions sp 
	JOIN sales s on s.id = sp.sale_id
	WHERE s.customer_id = $1;
	`,id)
	if err != nil {
		return nil, ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		sale := &Sales{}
		err = rows.Scan(&sale.ID, &sale.Name, &sale.Price, &sale.Qty, &sale.Created)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		sales = append(sales, sale)
	}
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil,err
	}
	return sales, nil
}  

// method for generating a token
func (s *Service) Token(ctx context.Context, phone string, password string) (token string, err error) {
	var hash string
	var id int64

	err = s.pool.QueryRow(ctx, `SELECT id,password FROM customers WHERE phone =$1`, phone).Scan(&id, &hash)

	if err == pgx.ErrNoRows {
		return "", ErrNoSuchUser
	}
	if err != nil {
		return "", ErrInternal
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return "", ErrInvalidPassword
	}

	buffer := make([]byte, 256)

	n, err := rand.Read(buffer)

	if n != len(buffer) || err != nil {
		return "", ErrInternal
	}

	token = hex.EncodeToString(buffer)
	_, err = s.pool.Exec(ctx, `INSERT INTO customers_tokens(token,customer_id) VALUES($1,$2)`, token, id)
	if err != nil {
		return "", ErrInternal
	}

	return token, nil
}



func (s *Service) ByID(ctx context.Context, id int64) (*Customer, error) {
	item := &Customer{}

	err := s.pool.QueryRow(ctx, `
		SELECT id, name, phone, active, created FROM customers WHERE id = $1
	`, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return item, nil

}

func (s *Service) All(ctx context.Context) (items []*Customer, err error) {

	rows, err := s.pool.Query(ctx, `
		SELECT * FROM customers
	`)

	for rows.Next() {
		item := &Customer{}
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Phone,
			&item.Active,
			&item.Created)
		if err != nil {
			log.Print(err)
		}

		items = append(items, item)
	}
	return items, nil
}
func (s *Service) AllActive(ctx context.Context) (items []*Customer, err error) {

	rows, err := s.pool.Query(ctx, `
		SELECT * FROM customers WHERE active
	`)

	for rows.Next() {
		item := &Customer{}
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Phone,
			&item.Active,
			&item.Created)
		if err != nil {
			log.Print(err)
		}

		items = append(items, item)
	}
	return items, nil
}

// //Save method
func (s *Service) Save(ctx context.Context, customer *Customer) (c *Customer, err error) {

	item := &Customer{}

	if customer.ID == 0 {
		sqlStatement := `insert into customers(name, phone, password) values($1, $2, $3) returning *`
		err = s.pool.QueryRow(ctx, sqlStatement, customer.Name, customer.Phone, customer.Password).Scan(
			&item.ID,
			&item.Name,
			&item.Phone,
			&item.Password,
			&item.Active,
			&item.Created)
	} else {
		sqlStatement := `update customers set name=$1, phone=$2, password=$3 where id=$4 returning *`
		err = s.pool.QueryRow(ctx, sqlStatement, customer.Name, customer.Phone, customer.Password, customer.ID).Scan(
			&item.ID,
			&item.Name,
			&item.Phone,
			&item.Password,
			&item.Active,
			&item.Created)
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return item, nil

}

func (s *Service) RemoveById(ctx context.Context, id int64) (*Customer, error) {
	item := &Customer{}
	err := s.pool.QueryRow(ctx, `
	DELETE FROM customers WHERE id=$1 RETURNING id,name,phone,active,created
	`, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return item, nil

}

func (s *Service) BlockByID(ctx context.Context, id int64) (*Customer, error) {
	item := &Customer{}
	err := s.pool.QueryRow(ctx, `
		UPDATE customers SET active = false WHERE id = $1 RETURNING id, name, phone, active, created
	`, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return item, nil

}
func (s *Service) UnBlockByID(ctx context.Context, id int64) (*Customer, error) {
	item := &Customer{}
	err := s.pool.QueryRow(ctx, `
		UPDATE customers SET active = true WHERE id = $1 RETURNING id, name, phone, active, created
	`, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return item, nil

}
