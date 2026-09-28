package models

import (
	"database/sql"

	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/database"
)

// RadiusUser mirrors the Node shape: {id, user, password, profile}.
type RadiusUser struct {
	ID       int64   `json:"id"`
	User     string  `json:"user"`
	Password string  `json:"password"`
	Profile  *string `json:"profile"`
}

func scanUser(row interface {
	Scan(dest ...interface{}) error
}) (*RadiusUser, error) {
	var u RadiusUser
	var profile sql.NullString
	if err := row.Scan(&u.ID, &u.User, &u.Password, &profile); err != nil {
		return nil, err
	}
	if profile.Valid {
		s := profile.String
		u.Profile = &s
	}
	return &u, nil
}

const userSelect = `
	SELECT rc.id as id, rc.username as user, rc.value as password, rr.value as profile
	FROM radcheck rc
	LEFT JOIN radreply rr ON rc.username = rr.username AND rr.attribute = 'Mikrotik-Group'
	WHERE rc.attribute = 'Cleartext-Password'`

func GetAllUsers(search string) ([]*RadiusUser, error) {
	q := userSelect + " ORDER BY rc.username"
	args := []interface{}{}
	if search != "" {
		q = userSelect + " AND rc.username LIKE ? ORDER BY rc.username"
		args = append(args, "%"+search+"%")
	}
	rows, err := database.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*RadiusUser{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func GetUserByUsername(username string) (*RadiusUser, error) {
	row := database.DB.QueryRow(userSelect+" AND rc.username = ?", username)
	u, err := scanUser(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func GetUserByID(id int64) (*RadiusUser, error) {
	row := database.DB.QueryRow(userSelect+" AND rc.id = ?", id)
	u, err := scanUser(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func UserExists(username string) (bool, error) {
	var u string
	err := database.DB.QueryRow(
		"SELECT username FROM radcheck WHERE username = ? AND attribute = 'Cleartext-Password'", username).Scan(&u)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func CreateUser(username, password, profile string) (*RadiusUser, error) {
	if profile == "" {
		profile = "PPP"
	}
	var id int64
	err := database.WithTx(func(tx *sql.Tx) error {
		res, err := tx.Exec("INSERT INTO radcheck (username, attribute, op, value) VALUES (?, 'Cleartext-Password', ':=', ?)", username, password)
		if err != nil {
			return err
		}
		id, _ = res.LastInsertId()
		_, err = tx.Exec("INSERT INTO radreply (username, attribute, op, value) VALUES (?, 'Mikrotik-Group', ':=', ?)", username, profile)
		return err
	})
	if err != nil {
		return nil, err
	}
	u, err := GetUserByUsername(username)
	if err == nil && u == nil {
		u = &RadiusUser{ID: id, User: username, Password: password, Profile: &profile}
	}
	return u, err
}

func UpdateUser(username, password, profile string, hasPassword, hasProfile bool) (*RadiusUser, error) {
	updated := false
	err := database.WithTx(func(tx *sql.Tx) error {
		if hasPassword {
			res, err := tx.Exec("UPDATE radcheck SET value = ? WHERE username = ? AND attribute = 'Cleartext-Password'", password, username)
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n > 0 {
				updated = true
			}
		}
		if hasProfile {
			var id int64
			err := tx.QueryRow("SELECT id FROM radreply WHERE username = ? AND attribute = 'Mikrotik-Group'", username).Scan(&id)
			if err == sql.ErrNoRows {
				_, err = tx.Exec("INSERT INTO radreply (username, attribute, op, value) VALUES (?, 'Mikrotik-Group', ':=', ?)", username, profile)
				if err != nil {
					return err
				}
				updated = true
			} else if err != nil {
				return err
			} else {
				res, err := tx.Exec("UPDATE radreply SET value = ? WHERE username = ? AND attribute = 'Mikrotik-Group'", profile, username)
				if err != nil {
					return err
				}
				if n, _ := res.RowsAffected(); n > 0 {
					updated = true
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !updated {
		return nil, nil
	}
	return GetUserByUsername(username)
}

func UpdateUserByID(id int64, password, profile string, hasPassword, hasProfile bool) (*RadiusUser, error) {
	var username string
	err := database.DB.QueryRow("SELECT username FROM radcheck WHERE id = ? AND attribute = 'Cleartext-Password'", id).Scan(&username)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u, err := UpdateUser(username, password, profile, hasPassword, hasProfile)
	if err != nil || u == nil {
		return u, err
	}
	return GetUserByID(id)
}

func DeleteUser(username string) (bool, error) {
	deleted := false
	err := database.WithTx(func(tx *sql.Tx) error {
		res, err := tx.Exec("DELETE FROM radcheck WHERE username = ?", username)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			deleted = true
		}
		_, err = tx.Exec("DELETE FROM radreply WHERE username = ?", username)
		return err
	})
	return deleted, err
}

func CountUsers() (int64, error) {
	var c int64
	err := database.DB.QueryRow("SELECT COUNT(DISTINCT username) FROM radcheck WHERE attribute = 'Cleartext-Password'").Scan(&c)
	return c, err
}

type Attribute struct {
	Attribute string `json:"attribute"`
	Op        string `json:"op"`
	Value     string `json:"value"`
}

func GetUserAttributes(username string) ([]Attribute, error) {
	rows, err := database.DB.Query("SELECT attribute, op, value FROM radcheck WHERE username = ?", username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Attribute{}
	for rows.Next() {
		var a Attribute
		if err := rows.Scan(&a.Attribute, &a.Op, &a.Value); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func GetUserReplyAttributes(username string) ([]Attribute, error) {
	rows, err := database.DB.Query("SELECT attribute, op, value FROM radreply WHERE username = ?", username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Attribute{}
	for rows.Next() {
		var a Attribute
		if err := rows.Scan(&a.Attribute, &a.Op, &a.Value); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func AddAttribute(username, attribute, op, value, table string) (int64, error) {
	if table != "radcheck" && table != "radreply" {
		table = "radcheck"
	}
	res, err := database.DB.Exec("INSERT INTO "+table+" (username, attribute, op, value) VALUES (?, ?, ?, ?)",
		username, attribute, op, value)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func RemoveAttribute(username, attribute, table string) (bool, error) {
	if table != "radcheck" && table != "radreply" {
		table = "radcheck"
	}
	res, err := database.DB.Exec("DELETE FROM "+table+" WHERE username = ? AND attribute = ?", username, attribute)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
