package config

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SenderEmail  string
	SenderName   string
	AuthEmail    string
	AuthPassword string
}

var DefaultEmailSettings = EmailConfig{

	SMTPHost:    "smtp.gmail.com",                  //  SMTP server, e.g., Gmail
	SMTPPort:    587,                               // Standard SMTP port for Gmail
	SenderEmail: "ishika.gupta@supersixsports.com", //  app's official email
	SenderName:  "monefyapplication",               //  app's name or your name
	AuthEmail:   "ishika.gupta@supersixsports.com", // Same as SenderEmail if using the same credentials
	// AuthPassword: os.Getenv("GMAIL_PASSWORD"),
	AuthPassword: "avrt gxav iqpn deit",

	// AuthEmail: "monefyapplication@gmail.com", // Same as SenderEmail if using the same credentials

}
