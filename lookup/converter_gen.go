package lookup

// Hey, this code is generated. You know the drill: DO NOT EDIT

func fromRedirectStorageToRedirectResult(i RedirectStorage) RedirectResult {
	return RedirectResult{
		Code:      i.Code,
		URL:       i.URL,
		CreatedAt: i.CreatedAt,
	}
}
