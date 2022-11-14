package email

type Email struct {
	Date    string
	From    string
	To      string
	Subject string
	Body    string
}

type EmailList struct {
	Index  string  `json:"index"`
	Emails []Email `json:"records"`
}
