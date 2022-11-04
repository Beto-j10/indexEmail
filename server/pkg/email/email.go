package email

type Email struct {
	Date    string
	From    string
	To      string
	Subject string
	Body    string
}

type EmailList struct {
	Emails []Email `json:"records"`
}
