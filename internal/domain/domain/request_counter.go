package domain

type RequestType string

const (
	Internal RequestType = "internal"
	External             = "external"

	MaxRequestsCount = 2
)
