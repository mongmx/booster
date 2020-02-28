package auth

type Service interface {
	SignIn() (interface{}, error)
}

type Member struct {
	ID int64
}
