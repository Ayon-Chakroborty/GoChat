package models

import (
	"database/sql"
	"errors"
)

type Chatroom struct {
	ID       int
	Name     string
	User     string
	Private  bool
	AllUsers string
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

func (m *ChatroomModel) Get(chatroom, email string) (*Chatroom, error) {
	stmt := `SELECT * FROM chatrooms WHERE name = ? AND user = ?`

	row := m.DB.QueryRow(stmt, chatroom, email)
	cr := &Chatroom{}

	err := row.Scan(&cr.ID, &cr.Name, &cr.User, &cr.Private)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return cr, nil
}

func (m *ChatroomModel) GetAllChats(email string) ([]*Chatroom, error) {
	stmt := `SELECT id, name, user, private FROM chatrooms
	WHERE user = ?`

	rows, err := m.DB.Query(stmt, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chatrooms := []*Chatroom{}

	for rows.Next() {
		cr := &Chatroom{}
		err := rows.Scan(&cr.ID, &cr.Name, &cr.User, &cr.Private)
		if err != nil {
			return nil, err
		}
		chatrooms = append(chatrooms, cr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return chatrooms, nil
}

func (m *ChatroomModel) GetUsersInChatroom(chatroom string) ([]string, error) {
	stmt := `SELECT users.username FROM users
	INNER JOIN chatrooms ON chatrooms.user=users.email AND chatrooms.name=?`

	rows, err := m.DB.Query(stmt, chatroom)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := []string{}

	for rows.Next() {
		n := ""
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return names, nil
}
