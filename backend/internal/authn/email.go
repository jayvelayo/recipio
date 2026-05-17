package authn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type EmailSender struct {
	APIKey string
	From   string
	AppURL string
}

func (s EmailSender) SendVerificationEmail(toEmail, name, rawToken string) error {
	verifyURL := s.AppURL + "/auth/verify-email?token=" + rawToken
	body := map[string]any{
		"from":    s.From,
		"to":      []string{toEmail},
		"subject": "Verify your Recipio email",
		"html": fmt.Sprintf(
			`<p>Hi %s,</p><p>Click <a href="%s">here to verify your email address</a>.</p><p>This link expires in 24 hours.</p>`,
			name, verifyURL,
		),
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("resend API returned status %d", resp.StatusCode)
	}
	return nil
}
