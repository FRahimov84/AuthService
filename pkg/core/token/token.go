package token

import (
	"context"
	"errors"
	"github.com/FRahimov84/myJwt/pkg/jwt"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Service struct {
	secret []byte
}

type Payload struct {
	Id       int64    `json:"id"`
	Username string   `json:"username"`
	Exp      int64    `json:"exp"`
	Roles    []string `json:"roles"`
}

type RequestDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResponseDTO struct {
	Token string `json:"token"`
}

var ErrInvalidLogin = errors.New("invalid password")
var ErrInvalidPassword = errors.New("invalid password")
var ErrServerError = errors.New("server error")

func NewService(secret jwt.Secret) *Service {
	return &Service{secret: secret}
}

func (s *Service) Generate(context context.Context, request *RequestDTO, pool *pgxpool.Pool) (response ResponseDTO, err error) {
	// TODO: Go to DB & get user by login
	conn, err := pool.Acquire(context)
	if err != nil {
		return
	}
	defer conn.Release()
	var (
		hash string
		id   int64
		isAdmin bool
	)
	err = conn.QueryRow(context, `select id, password, admin from users where username = $1 and removed = FALSE;`, request.Username).Scan(&id, &hash, &isAdmin)

	if err != nil {
		err = ErrServerError
		return
	}

	//hash, err := bcrypt.GenerateFromPassword([]byte("hash"), bcrypt.DefaultCost)
	//log.Print(hash)

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(request.Password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		err = ErrInvalidPassword
		return
	}
	role := "User"
	if isAdmin {
		role = "Admin"
	}
	response.Token, err = jwt.Encode(Payload{
		Id:  id,
		Username: request.Username,
		Exp: time.Now().Add(time.Hour).Unix(),
		Roles: []string{role},
	}, s.secret)
	if err != nil {
		return ResponseDTO{}, ErrServerError
	}
	return
}