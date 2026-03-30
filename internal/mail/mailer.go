package mail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host        string
	Port        int
	User        string
	Pass        string
	From        string
	StartTLS    bool
	ImplicitTLS bool
	BaseURL     string
}

func FromEnv() (*Config, error) {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	if host == "" {
		return nil, fmt.Errorf("missing SMTP_HOST")
	}
	port := 587
	if p := strings.TrimSpace(os.Getenv("SMTP_PORT")); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	if from == "" {
		from = user
	}
	startTLS := strings.EqualFold(os.Getenv("SMTP_STARTTLS"), "1") || strings.EqualFold(os.Getenv("SMTP_STARTTLS"), "true")
	implicitTLS := strings.EqualFold(os.Getenv("SMTP_TLS"), "1") || strings.EqualFold(os.Getenv("SMTP_TLS"), "true")
	baseURL := strings.TrimSpace(os.Getenv("BASE_URL"))
	return &Config{Host: host, Port: port, User: user, Pass: pass, From: from, StartTLS: startTLS, ImplicitTLS: implicitTLS, BaseURL: baseURL}, nil
}

type Mailer struct {
	cfg *Config
}

func New(cfg *Config) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) SendInvoice(to, subject, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	var c *smtp.Client

	if m.cfg.ImplicitTLS {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: m.cfg.Host})
		if err != nil {
			return err
		}
		c, err = smtp.NewClient(conn, m.cfg.Host)
		if err != nil {
			return err
		}
	} else {
		conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			return err
		}
		cc, err := smtp.NewClient(conn, m.cfg.Host)
		if err != nil {
			return err
		}
		c = cc
		if m.cfg.StartTLS {
			if ok, _ := c.Extension("STARTTLS"); ok {
				_ = c.StartTLS(&tls.Config{ServerName: m.cfg.Host})
			}
		}
	}
	defer c.Close()

	if m.cfg.User != "" {
		if ok, _ := c.Extension("AUTH"); ok {
			auth := smtp.PlainAuth("", m.cfg.User, m.cfg.Pass, m.cfg.Host)
			if err := c.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err := c.Mail(m.cfg.From); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	msg := buildHTMLMessage(m.cfg.From, to, subject, htmlBody)
	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}
	return c.Quit()
}

func buildHTMLMessage(from, to, subject, htmlBody string) string {
	subject = strings.ReplaceAll(subject, "\n", " ")
	b := &strings.Builder{}
	fmt.Fprintf(b, "From: %s\r\n", from)
	fmt.Fprintf(b, "To: %s\r\n", to)
	fmt.Fprintf(b, "Subject: %s\r\n", subject)
	fmt.Fprintf(b, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(b, "Content-Type: text/html; charset=UTF-8\r\n")
	fmt.Fprintf(b, "Content-Transfer-Encoding: 8bit\r\n")
	fmt.Fprintf(b, "\r\n")
	b.WriteString(htmlBody)
	return b.String()
}
