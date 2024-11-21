package models

import (
	"database/sql"
	"errors"
)

type Chatroom struct {
	ID     int
	Name    string
	User    string
	Private bool
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

func (m *ChatroomModel) Get(chatroom, email string) (*Chatroom, error){
	stmt := `SELECT * FROM chatrooms WHERE name = ? AND user = ?`
	
	row := m.DB.QueryRow(stmt, chatroom, email)
	cr := &Chatroom{}

	err := row.Scan(&cr.ID, &cr.Name, &cr.User, &cr.Private)
	if err != nil{
		if errors.Is(err, sql.ErrNoRows){
			return nil, ErrNoRecord
		} else{
			return nil, err
		}
	}

	return cr, nil
}
