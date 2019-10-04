package gmail

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestGmail(t *testing.T) {
	t.Skip("")

	c, err := New(context.Background(), os.Getenv("GMAIL_USER"), os.Getenv("GMAIL_CREDENTIAL_PATH"), os.Getenv("GMAIL_REFRESH_TOKEN"))
	assert.NoError(t, err)

	rows, err := c.List("", 10)
	assert.NoError(t, err)

	fmt.Printf("# msgs = %v\n", len(rows))
	for _, r := range rows {
		msg, err := c.ReadMessage(r.Id)
		assert.NoError(t, err)
		fmt.Printf("-----------------------------------\n")
		spew.Dump(msg)
	}
}
