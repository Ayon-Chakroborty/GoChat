package models

import (
	"database/sql"
)

type Chatroom struct {
	id      int
	name    string
	user    string
	private bool
}

type ChatroomModel struct {
	DB *sql.DB
}

func (m *ChatroomModel) Insert(name string, user string, private bool) error {
	stmt := `INSERT INTO chatrooms (name, user, private) VALUES (?, ?, ?)`

	_, err := m.DB.Exec(stmt, name, user, private)
	if err != nil {
		return err
	}

	return nil
}
