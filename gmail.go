package gmail

import (
	"bytes"
	"context"
	"encoding/base64"
	"io/ioutil"
	"mime"
	"mime/quotedprintable"
	"strings"
	"time"

	"github.com/jordan-wright/email"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gm "google.golang.org/api/gmail/v1"
)

type Client struct {
	ctx  context.Context
	user string
	s    *gm.Service
	ums  *gm.UsersMessagesService
}

// Refer to https://stackoverflow.com/questions/37534548/how-to-access-a-gmail-account-i-own-using-gmail-api
// We need to create OAuth client in google api console
// (https://console.developers.google.com/apis). Then use
// the client id and secret to generate refresh token in
// auth playground (https://developers.google.com/oauthplayground)
// Download the corresonding credentials json file as well.
// Note: when creating OAuth client, set type as web application,
// and use "https://developers.google.com/oauthplayground" as redirect URI.
func New(ctx context.Context, user, credentialPath, refreshToken string) (*Client, error) {
	b, err := ioutil.ReadFile(credentialPath)
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gm.MailGoogleComScope)
	if err != nil {
		return nil, err
	}

	// Use access token and refresh token to regenerate a new
	// access token with expiration.
	token := &oauth2.Token{
		AccessToken:  "",
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-time.Hour),
	}
	if token, err = config.TokenSource(ctx, token).Token(); err != nil {
		return nil, err
	}

	s, err := gm.New(config.Client(ctx, token))
	if err != nil {
		return nil, err
	}

	return &Client{
		ctx:  ctx,
		user: user,
		s:    s,
		ums:  gm.NewUsersMessagesService(s),
	}, nil
}

type Row struct {
	Id       string
	ThreadId string
}

func (c *Client) List(q string, n int64) ([]*Row, error) {
	umlc := c.ums.List(c.user).IncludeSpamTrash(true).Q(q)
	if n > 0 {
		umlc = umlc.MaxResults(n)
	}
	var rows []*Row
	if err := umlc.Pages(c.ctx, func(resp *gm.ListMessagesResponse) error {
		for _, m := range resp.Messages {
			rows = append(rows, &Row{
				Id:       m.Id,
				ThreadId: m.ThreadId,
			})
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return rows, nil
}

type Message struct {
	Id      string
	ReplyTo []string
	From    string
	To      []string
	Bcc     []string
	Cc      []string
	Subject string
	Text    string
	HTML    string
	Sender  string
	Date    time.Time
}

func decode(s string) string {
	if strings.HasPrefix(s, "=?") && strings.HasSuffix(s, "?=") {
		// =?UTF-8?q?=E9=82=AE=E4=BB=B6=E9=AA=8C=E8=AF=81=E7=A0=81?=
		dec := &mime.WordDecoder{}
		if decoded, err := dec.Decode(s); err != nil {
			return s
		} else {
			return decoded
		}
	}

	if decoded, err := ioutil.ReadAll(quotedprintable.NewReader(strings.NewReader(s))); err == nil {
		return string(decoded)
	}

	return s
}

func decodeSlice(s []string) []string {
	ret := []string{}
	for _, str := range s {
		ret = append(ret, decode(str))
	}
	return ret
}

func (c *Client) ReadMessage(id string) (*Message, error) {
	resp, err := c.ums.Get(c.user, id).Format("RAW").Do()
	if err != nil {
		return nil, err
	}

	decoded, err := base64.URLEncoding.DecodeString(resp.Raw)
	if err != nil {
		return nil, err
	}

	m, err := email.NewEmailFromReader(bytes.NewReader(decoded))
	var date time.Time
	dateStr := m.Headers.Get("Date")
	if dateStr != "" {
		// Tue, 9 Jul 2019 14:46:08 +0800
		// Tue, 9 Jul 2019 14:46:08 +0000 (UTC)
		parts := strings.Split(dateStr, " (")
		if date, err = time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", parts[0]); err != nil {
			return nil, err
		}
	}

	return &Message{
		Id:      id,
		ReplyTo: decodeSlice(m.ReplyTo),
		From:    decode(m.From),
		To:      decodeSlice(m.To),
		Bcc:     decodeSlice(m.Bcc),
		Cc:      decodeSlice(m.Cc),
		Subject: decode(m.Subject),
		Text:    decode(string(m.Text)),
		HTML:    decode(string(m.HTML)),
		Sender:  m.Sender,
		Date:    date,
	}, nil
}
