package gp_pg

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"strings"
)

var (
	Hostname = ""
	Port     = 2345
	Username = ""
	Password = ""
	Database = ""
)

type Userdata struct {
	ID          int
	Username    string
	Name        string
	Surname     string
	Description string
}

func openConnection() (*sql.DB, error) {
	conn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", Hostname, Port, Username, Password, Database)
	db, err := sql.Open("posgresql", conn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func exists(username string) int {
	username = strings.ToLower(username)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	userID := -1
	statement := fmt.Sprintf(`SELECT id FROM users WHERE username=%s`, username)
	rows, err := db.Query(statement)

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			fmt.Println(err)
			return -1
		}
		userID = id
	}
	defer rows.Close()

	return userID
}

func AddUser(d Userdata) int {
	d.Username = strings.ToLower(d.Username)
	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	userID := exists(d.Username)
	if userID != -1 {
		fmt.Printf("User %s already exists\n", d.Username)
		return -1
	}

	insertStatement := `INSERT INTO user (username) VALUES ($1)`

	_, err = db.Exec(insertStatement, d.Username)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	userID = exists(d.Username)
	if userID == -1 {
		return userID
	}

	insertStatement = `INSERT INTO userdata (userid, name, surname, description) VALUES ($1, $2, $3, $4)`

	_, err = db.Exec(insertStatement, userID, d.Name, d.Surname, d.Description)
	if err != nil {
		fmt.Println("db.Exec()", err)
		return -1
	}

	return userID
}

func DeleteUser(id int) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.CLose()
	
	statement := fmt.Sprintf("SELECT username FROM users WHERE id=%d", id)
	rows, err := db.Query(statement)
	var username string
	for rows.Next() {
		err = rows.Scan(&username)
		if err != nil {
			return err
		}
	}
	defer rows.Close()
	
	if exists(username) != id {
		fmt.Errorf("User with ID %d does not exist\n", id)
	}
	
	deleteStatement := `DELETE FROM userdata WHERE userid=$1`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}
	
	deleteStatement = `DELETE FROM users WHERE id=$1`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}
	
	return nil
}

func ListUsers() ([]Userdata, error) {
	data := []Userdata{}
	db, err := openConnection()
	if err != nil {
		return data, err
	}
	defer db.Close()
	
	rows, err := db.Query(`SELECT id, username, surname, description FROM users, userdata WHERE users.id = userdata.id`)
	if err != nil {
		return data, err
	}
	for rows.Next() {
		userDataTemp := Userdata{}
		err = rows.Scan(&userDataTemp.ID, &userDataTemp.Username, &userDataTemp.Surname, &userDataTemp.Description)
		
		data = append(data, userDataTemp)
		if err != nil {
			return data, err
		}
	}
	defer rows.Close()
	
	return data, nil
}

func UpdateUser(d Userdata) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	userID := exists(d.Username)
	if userID == -1 {
		return errors.New("User with username does not exist")
	}

	d.ID = userID
	updateStatement := `UPDATE userdata set name=$1, surname=$2, description=$3 where userid=$4`
	err = db.Exec(updateStatement, d.Username, d.Surname, d.Description, d.ID)
	if err != nil {
		return err
	}

	return nil
}