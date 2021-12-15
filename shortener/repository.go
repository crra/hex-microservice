package shortener

type Repository interface {
	Find(code string) (*Redirect, error)
	Store(redirect *Redirect) error
}
