package memory

// Hey, this code is generated. You know the drill: DO NOT EDIT

import (
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/lookup"
)

func fromRedirectToLookupRedirectStorage(i redirect) lookup.RedirectStorage {
	return lookup.RedirectStorage{
		Code:      i.Code,
		URL:       i.URL,
		CreatedAt: i.CreatedAt,
	}
}

func fromAdderRedirectStorageToRedirect(i adder.RedirectStorage) redirect {
	return redirect{
		Code:      i.Code,
		URL:       i.URL,
		Token:     i.Token,
		CreatedAt: i.CreatedAt,
	}
}

func fromRedirectToDeleterRedirectStorage(i redirect) deleter.RedirectStorage {
	return deleter.RedirectStorage{
		Code:  i.Code,
		Token: i.Token,
	}
}
