package adder

import (
	"fmt"

	"github.com/jinzhu/copier"
)

//
// This is generated code, do not edit
//

func storageToResult(stored RedirectStorage) (RedirectResult, error) {
	var red RedirectResult

	if err := copier.Copy(&red, &stored); err != nil {
		return red, fmt.Errorf("storageToResult copying: %w", err)
	}

	return red, nil
}
