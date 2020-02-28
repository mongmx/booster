package member

// Repository is the domain storage
type Repository interface {
}

type Member struct {
	ID int64 `json:"id"`
}
