package service

type SearchAttributeCriteria struct {
	SearchField string
}

type SearchUsersCriteria struct {
	SearchField   string
	Email         string
	AttributeKeys []string
}
