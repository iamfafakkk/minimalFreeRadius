package models

import (
	"database/sql"

	"github.com/iamfafakkk/minimalFreeRadius/freeradius-api/internal/database"
)

// NAS uses API field names (name/ip) mapped to DB columns (shortname/nasname).
type NAS struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	IP          string `json:"ip"`
	Secret      string `json:"secret"`
	Type        string `json:"type"`
	Ports       *int   `json:"ports"`
	Community   string `json:"community"`
	Description string `json:"description"`
}

func scanNAS(row interface {
	Scan(dest ...interface{}) error
}) (*NAS, error) {
	var n NAS
	var typ, community, desc sql.NullString
	var ports sql.NullInt64
	var secret sql.NullString
	if err := row.Scan(&n.ID, &n.Name, &n.IP, &secret, &typ, &ports, &community, &desc); err != nil {
		return nil, err
	}
	if secret.Valid {
		n.Secret = secret.String
	}
	if typ.Valid {
		n.Type = typ.String
	}
	if ports.Valid {
		p := int(ports.Int64)
		n.Ports = &p
	}
	if community.Valid {
		n.Community = community.String
	}
	if desc.Valid {
		n.Description = desc.String
	}
	return &n, nil
}

const nasCols = "id, shortname, nasname, secret, type, ports, community, description"

func GetAllNAS(search string) ([]*NAS, error) {
	var rows *sql.Rows
	var err error
	if search != "" {
		pat := "%" + search + "%"
		rows, err = database.DB.Query(
			"SELECT "+nasCols+" FROM nas WHERE shortname LIKE ? OR nasname LIKE ? OR description LIKE ? ORDER BY shortname",
			pat, pat, pat)
	} else {
		rows, err = database.DB.Query("SELECT " + nasCols + " FROM nas ORDER BY shortname")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*NAS{}
	for rows.Next() {
		n, err := scanNAS(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func GetNASByID(id int64) (*NAS, error) {
	row := database.DB.QueryRow("SELECT "+nasCols+" FROM nas WHERE id = ?", id)
	n, err := scanNAS(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return n, err
}

func NASExists(shortname, nasname string, excludeID *int64) (bool, error) {
	q := "SELECT id FROM nas WHERE (shortname = ? OR nasname = ?)"
	args := []interface{}{shortname, nasname}
	if excludeID != nil {
		q += " AND id != ?"
		args = append(args, *excludeID)
	}
	var id int64
	err := database.DB.QueryRow(q, args...).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

type NASCreate struct {
	Name, IP, Secret, Type, Community, Description string
	Ports                                          int
}

func CreateNAS(in NASCreate) (*NAS, error) {
	res, err := database.DB.Exec(
		"INSERT INTO nas (shortname, nasname, secret, type, ports, community, description) VALUES (?, ?, ?, ?, ?, ?, ?)",
		in.Name, in.IP, in.Secret, in.Type, in.Ports, in.Community, in.Description)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return GetNASByID(id)
}

func UpdateNAS(id int64, cur *NAS, name, ip, secret, typ, community, desc string, ports int) (*NAS, error) {
	if name == "" {
		name = cur.Name
	}
	if ip == "" {
		ip = cur.IP
	}
	if secret == "" {
		secret = cur.Secret
	}
	if typ == "" {
		typ = cur.Type
	}
	if ports == 0 {
		if cur.Ports != nil {
			ports = *cur.Ports
		} else {
			ports = 1812
		}
	}
	res, err := database.DB.Exec(
		"UPDATE nas SET shortname=?, nasname=?, secret=?, type=?, ports=?, community=?, description=? WHERE id=?",
		name, ip, secret, typ, ports, community, desc, id)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, nil
	}
	return GetNASByID(id)
}

func DeleteNAS(id int64) (bool, error) {
	res, err := database.DB.Exec("DELETE FROM nas WHERE id = ?", id)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func CountNAS() (int64, error) {
	var c int64
	err := database.DB.QueryRow("SELECT COUNT(*) FROM nas").Scan(&c)
	return c, err
}
