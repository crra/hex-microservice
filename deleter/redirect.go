package deleter

// RedirectStorage is the storage view for the lookup service.
type RedirectStorage struct {
	Code  string
	Token string
}

// RedirectQuery is the request query of the deleter service.
type RedirectQuery struct {
	Code  string
	Token string
}
