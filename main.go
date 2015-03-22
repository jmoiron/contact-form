package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"time"
)

const spamcheckUrl = "http://spamcheck.postmarkapp.com/filter"
const fallbackHost = "unknown.host.com"

var cfg struct {
	port      int
	nospam    bool
	mailhost  string
	mailuser  string
	mailpass  string
	mailport  int
	destemail string
}

func init() {
	EnvInt(&cfg.port, "CONTACT_PORT", 3241)
	EnvBool(&cfg.nospam, "CONTACT_NOSPAM", false)
	EnvInt(&cfg.mailport, "CONTACT_MAILPORT", 587)
	EnvString(&cfg.mailhost, "CONTACT_MAILHOST", "smtp.google.com")
	EnvString(&cfg.mailuser, "CONTACT_MAILUSER", "")
	EnvString(&cfg.mailpass, "CONTACT_MAILPASS", "")
	EnvString(&cfg.destemail, "CONTACT_DESTEMAIL", "")

	// we want to tread lightly with showing the environ's password potentially
	// by accident when running the -help, so lets show something else instead
	defaultPass := "$CONTACT_MAILPASS"
	if len(cfg.mailpass) != 0 {
		defaultPass += " (SET)"
	}

	var flagPass string

	flag.IntVar(&cfg.port, "port", cfg.port, "http port")
	flag.BoolVar(&cfg.nospam, "nospam", cfg.nospam, "disable spam check")
	flag.IntVar(&cfg.mailport, "mailport", cfg.mailport, "port to send mail on")
	flag.StringVar(&cfg.mailhost, "mailhost", cfg.mailhost, "host to send mail from")
	flag.StringVar(&cfg.mailuser, "mailuser", cfg.mailuser, "username for mailhost")
	flag.StringVar(&flagPass, "mailpass", defaultPass, "password for mailhost")
	flag.StringVar(&cfg.destemail, "destemail", cfg.destemail, "destination mailbox")
	flag.Parse()

	if flagPass != defaultPass {
		cfg.mailpass = flagPass
	}
}

func main() {
	fs := http.FileServer(http.Dir("."))
	http.HandleFunc("/contact/", contact)
	http.Handle("/", fs)
	fmt.Printf("Listening on :%d\n", cfg.port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.port), nil)
}

type msi map[string]interface{}

// contact handles a contact form submission.
func contact(w http.ResponseWriter, r *http.Request) {
	// we are always responding with a json message
	w.Header().Add("Content-Type", "application/json")

	msg := &Message{
		From:    r.FormValue("from"),
		Subject: r.FormValue("subject"),
		Body:    r.FormValue("body"),
	}

	if !msg.Validate() {
		msg.Errors["success"] = false
		json.NewEncoder(w).Encode(msg.Errors)
		return
	}

	logJson(msg)

	if !msg.SpamCheck() {
		msg.Errors["success"] = false
		msg.Errors["form"] = "Sorry, this failed a spam check"
		json.NewEncoder(w).Encode(msg.Errors)
		return
	}

	err := msg.Deliver()
	if err != nil {
		log.Printf("Error delivering email: %s", err)
		json.NewEncoder(w).Encode(msi{
			"success": false,
			"form":    "Sorry, there was a problem sending email",
		})
		return
	}

	json.NewEncoder(w).Encode(msi{"success": true})
}

func logJson(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Println(err)
	}
	log.Println(string(b))
}

func messageId() string {
	host, err := os.Hostname()
	if err != nil {
		host = fallbackHost
	}
	return fmt.Sprintf("<%d@%s>", rand.Int(), host)
}

type Message struct {
	From    string                 `json:"from"`
	Subject string                 `json:"subject"`
	Body    string                 `json:"body"`
	Errors  map[string]interface{} `json:"errors"`
}

func (m *Message) FullBody() string {
	var buf bytes.Buffer
	date := time.Now().Local().Format(time.RFC822Z)
	buf.WriteString(fmt.Sprintf("From: %s\r\n", m.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", cfg.destemail))
	buf.WriteString(fmt.Sprintf("Reply-To: %s\r\n", m.From))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", date))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", m.Subject))
	buf.WriteString(fmt.Sprintf("Message-Id: %s\r\n", messageId()))
	buf.WriteString(m.Body)
	return buf.String()
}

// Validate returns whether or not this message is valid.
func (m *Message) Validate() bool {
	m.Errors = make(map[string]interface{})

	re := regexp.MustCompile(".+@.+\\..+")
	matched := re.Match([]byte(m.From))

	if matched == false {
		m.Errors["from"] = "Please enter a valid email address"
	}

	if strings.TrimSpace(m.Body) == "" {
		m.Errors["body"] = "Please write a message"
	}

	return len(m.Errors) == 0
}

// SpamCheck uses the postmarkapp API to check a message for spam.
func (m *Message) SpamCheck() bool {
	// if we should skip the spam check, just always return true
	if cfg.nospam {
		return true
	}
	// perform a spam check using the postmark API:
	// http://spamcheck.postmarkapp.com/doc
	req := SpamCheckReq{Email: m.FullBody(), Options: "long"}
	resp, err := req.Post()
	if err != nil {
		log.Printf("Error with spamcheck: %s", err)
		return true
	}
	logJson(resp)
	return resp.Success
}

// Deliver this message.
func (m *Message) Deliver() error {
	to := []string{cfg.destemail}
	body := []byte(m.FullBody())
	sendaddr := fmt.Sprintf("%s:%d", cfg.mailhost, cfg.mailport)

	auth := smtp.PlainAuth("", cfg.mailuser, cfg.mailpass, cfg.mailhost)
	log.Printf("Sending mail to %s from %s to %s", sendaddr, m.From, to)
	return smtp.SendMail(sendaddr, auth, m.From, to, body)
}

type SpamCheckResp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Report  string `json:"report"`
	Score   string `json:"score"`
	resp    *http.Response
}

type SpamCheckReq struct {
	Email   string `json:"email"`
	Options string `json:"options"`
}

func (s *SpamCheckReq) Post() (*SpamCheckResp, error) {
	body, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	hresp, err := http.Post(spamcheckUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	resp := &SpamCheckResp{resp: hresp}
	err = json.NewDecoder(hresp.Body).Decode(resp)
	return resp, err
}
