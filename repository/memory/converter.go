package memory

//
// This is generated code, do not edit
//

import (
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/lookup"

	"github.com/jinzhu/copier"
)

// TODO: replace with generated

func redirectToLookupRedirectStorage(stored redirect) (lookup.RedirectStorage, error) {
	var red lookup.RedirectStorage

	if err := copier.Copy(&red, &stored); err != nil {
		return red, fmt.Errorf("redirectToLookupRedirectStorage copying: %w", err)
	}

	return red, nil
}

func adderRedirectStorageToRedirect(red adder.RedirectStorage) (redirect, error) {
	var toStore redirect

	if err := copier.Copy(&toStore, &red); err != nil {
		return toStore, fmt.Errorf("dderRedirectStorageToRedirect copying: %w", err)
	}

	return toStore, nil
}

func redirectToDeleterRedirectStorage(stored redirect) (deleter.RedirectStorage, error) {
	var red deleter.RedirectStorage

	if err := copier.Copy(&red, &stored); err != nil {
		return red, fmt.Errorf("redirectToDeleterRedirectStorage copying: %w", err)
	}

	return red, nil
}
