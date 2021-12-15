package shortener

type Service interface {
	Find(code string) (*Redirect, error)
	Store(redirect *Redirect) error
}
