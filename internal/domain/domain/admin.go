package domain

type UsersReport struct {
	TotalInfo TotalInfoForReport
	UsersInfo []UserInfoForReport
}

type TotalInfoForReport struct {
	UsersAmount int64
	RPM         int64
	FPM         int64
	GPM         float64
}

type UserInfoForReport struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Plan      string
	GPM       float64
	RPM       int64
	Feedbacks []string
}
