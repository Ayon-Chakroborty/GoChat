package models

import (
	"database/sql"
	"time"
)

type Chat struct {
	ID       int
	Chatroom string
	Sender   string
	Private  bool
	Message  string
	Created  time.Time
	Username string
}

type ChatModel struct {
	DB *sql.DB
}

func (m *ChatModel) Insert(chatroom string, sender string, private bool, message string, username string) error {
	stmt := `INSERT INTO chats (chatroom, sender, private, message, created, username) 
	VALUES (?, ?, ?, ?, UTC_TIMESTAMP(), ?)`

	_, err := m.DB.Exec(stmt, chatroom, sender, private, message, username)
	if err != nil {
		return err
	}

	return nil
}

func (m *ChatModel) Get(chatroom string) ([]*Chat, error) {
	stmt := `SELECT * FROM chats WHERE chatroom = ?
	ORDER BY created ASC LIMIT 200`
	rows, err := m.DB.Query(stmt, chatroom)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chats := []*Chat{}
	for rows.Next() {
		c := &Chat{}
		err = rows.Scan(&c.ID, &c.Chatroom, &c.Sender, &c.Private, &c.Message, &c.Created, &c.Username)
		if err != nil {
			return nil, err
		}
		chats = append(chats, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

func (m *ChatModel) DeleteUser(email string) (error){
	stmt := `DELETE FROM chats WHERE sender=?`

	_, err := m.DB.Exec(stmt, email)
	if err != nil{
		return err
	}

	return nil
}