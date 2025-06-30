package service

type SearchAttributeCriteria struct {
	SearchField string
}

type SearchAuthorityCriteria struct {
	SearchField string
}

type SearchUserCriteria struct {
	SearchField   string
	Email         string
	AttributeKeys []string
}
