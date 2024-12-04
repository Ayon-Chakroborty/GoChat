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

func (m *ChatroomModel) Get(chatroom string, email string, private bool) (*Chatroom, error) {
	stmt := `SELECT * FROM chatrooms WHERE name = ? AND user = ? and private = ?`

	row := m.DB.QueryRow(stmt, chatroom, email, private)
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

func (m *ChatroomModel) Delete(chatroom, email string) error{
	stmt := `DELETE FROM chatrooms WHERE name=? AND user=?`

	_, err := m.DB.Exec(stmt, chatroom, email)
	if err != nil{
		return err
	}

	return nil
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

func (m *ChatroomModel) GetUsersInChatroom(chatroom string, private bool) ([]string, error) {
	stmt := `SELECT users.username FROM users
	INNER JOIN chatrooms ON chatrooms.user=users.email AND chatrooms.name=? AND chatrooms.private=?`

	rows, err := m.DB.Query(stmt, chatroom, private)
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

func (m *ChatroomModel) DeleteUser(email string) error {
	stmt := `DELETE FROM chatrooms WHERE user=?`

	_, err := m.DB.Exec(stmt, email)
	if err != nil {
		return err
	}

	return nil
}

func (m *ChatroomModel) SearchUser(email, searchEmail string) ([]*Chatroom, error) {
	stmt := `select id, name, user, private from chatrooms 
	where name in(select name from chatrooms 
	where user in (?, ?) group by name having count(distinct user)=2) and user=?`

	rows, err := m.DB.Query(stmt, email, searchEmail, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chatrooms := []*Chatroom{}

	for rows.Next() {
		cr := &Chatroom{}
		if err := rows.Scan(&cr.ID, &cr.Name, &cr.User, &cr.Private); err != nil{
			return nil, err
		}
		chatrooms = append(chatrooms, cr)
	}

	if err = rows.Err(); err != nil{
		return nil, err
	}

	return chatrooms, nil
}

func (m *ChatroomModel) GetSearchedChat(email, chatroom string) ([]*Chatroom, error) {
	stmt := `SELECT id, name, user, private FROM chatrooms
	WHERE user=? AND name=?`

	rows, err := m.DB.Query(stmt, email, chatroom)
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
