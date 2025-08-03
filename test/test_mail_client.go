package test

import client2 "github.com/janobono/auth-service/internal/service/client"

type testMailClient struct {
}

func (t testMailClient) SendEmail(data *client2.MailData) {
}
