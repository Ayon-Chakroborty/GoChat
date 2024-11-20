package models

import (
	"database/sql"
	"time"
)

type Chat struct {
	id       int
	chatroom string
	sender   string
	private  bool
	message  string
	created  time.Time
}

type ChatModel struct {
	DB *sql.DB
}

func (m *ChatModel) Insert(chatroom string, sender string, private bool, message string) error {
	stmt := `INSERT INTO chats (chatroom, sender, private, message, created) 
	VALUES (?, ?, ?, ?, UTC_TIMESTAMP())`

	_, err := m.DB.Exec(stmt, chatroom, sender, private, message)
	if err != nil {
		return err
	}

	return nil
}
