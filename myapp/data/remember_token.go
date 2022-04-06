package data

import (
	"fmt"
	"time"

	up "github.com/upper/db/v4"
)

type RememberToken struct {
	ID            int       `db:"id,omitempty"`
	UserID        int       `db:"user_id"`
	RememberToken string    `db:"remember_token"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func (t *RememberToken) Table() string {
	return "remember_tokens"
}

func (t *RememberToken) InsertToken(userID int, token string) error {
	logSnippet := "\n[remember_token.go][InsertToken] =>"

	collection := upper.Collection(t.Table())
	rememberToken := RememberToken{
		UserID:        userID,
		RememberToken: token,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	ir, err := collection.Insert(rememberToken)
	if err != nil {
		fmt.Printf("%s (Insert): %v", logSnippet, err)
		return err
	}
	fmt.Printf("%s (InsertResult.ID): '%d'", logSnippet, ir.ID())

	return nil
}

func (t *RememberToken) DeleteToken(rememberToken string) error {
	logSnippet := "\n[remember_token.go][DeleteToken] =>"

	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"remember_token": rememberToken})
	err := res.Delete()
	if err != nil {
		fmt.Printf("%s (Delete): %v", logSnippet, err)
		return err
	}

	return nil
}
