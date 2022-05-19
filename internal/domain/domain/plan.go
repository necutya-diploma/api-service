package domain

import "time"

const (
	BasicPlanName = "Basic"
)

type Plan struct {
	ID                    string
	Name                  string
	Description           string
	Options               []string
	Price                 int
	Duration              int
	InternalRequestsCount int
	ExternalRequestsCount int
}

func (p *Plan) IsBasic() bool {
	return p.Name == BasicPlanName
}

func (p *Plan) EndDateFromNow() *time.Time {
	if p.Duration == 0 {
		return nil
	}

	endDate := time.Now().AddDate(0, p.Duration, 0)

	return &endDate
}
