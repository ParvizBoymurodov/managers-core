package core

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

func TestLoginClient_QueryError(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()

	_,_, err = Login("", "", db)
	var typedErr *QueryError
	if ok := errors.As(err, &typedErr); !ok {
		t.Errorf("error not maptch QueryError: %v", err)
	}
}

func TestLoginClient_NoSuchLoginForEmptyDb(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()

	_, err = db.Exec(`
  CREATE TABLE client (
   id INTEGER PRIMARY KEY AUTOINCREMENT,
  login TEXT NOT NULL UNIQUE,
  password TEXT NOT NULL)`)
	if err != nil {
		t.Errorf("can't execute query: %v", err)
	}

	_, result, err := Login("", "", db)
	if err != nil {
		t.Errorf("can't execute Login: %v", err)
	}

	if result != false {
		t.Error("Login result not false for empty table")
	}
}


func TestLoginClient_LoginOk(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()


	_, err = db.Exec(`
  CREATE TABLE client (
   id INTEGER PRIMARY KEY AUTOINCREMENT,
  login TEXT NOT NULL UNIQUE,
  password TEXT NOT NULL)`)
	if err != nil {
		t.Errorf("can't execute Login: %v", err)
	}

	_, err = db.Exec(`insert into client (login,password) values ("vasya", "secret")`)
	if err != nil {
		t.Errorf("can't execute Login: %v", err)
	}

	_,result, err := Login("vasya", "secret" , db)
	if err != nil {
		t.Errorf("can't execute Login: %v", err)
	}

	if result != true {
		t.Error("Login result not true for existing account")
	}
}

func TestLoginClient_LoginNotOkForInvalidPassword(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()


	_, err = db.Exec(`
 CREATE TABLE client (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
 login TEXT NOT NULL UNIQUE,
 password TEXT NOT NULL)`)
	if err != nil {
		t.Errorf("can't execute Login: %v", err)
	}

	_, err = db.Exec(`INSERT INTO client(id, login, password) VALUES (1, 'vasya', 'secret')`)
	if err != nil {
		t.Errorf("can't execute Login: %v", err)
	}

	_,_, err = Login("vasya", "password", db)
	if !errors.Is(err, ErrInvalidPass) {
		t.Errorf("Not ErrInvalidPass error for invalid pass: %v", err)
	}
}


func TestLoginManager_QueryError(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()

	_, err = LoginForManagers("", "", db)
	var typedErr *QueryError
	if ok := errors.As(err, &typedErr); !ok {
		t.Errorf("error not maptch QueryError: %v", err)
	}
}

func TestLoginManager_NoSuchLoginForEmptyDb(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()

	_, err = db.Exec(`
  CREATE TABLE managers (
   id INTEGER PRIMARY KEY AUTOINCREMENT,
  login TEXT NOT NULL UNIQUE,
  password TEXT NOT NULL)`)
	if err != nil {
		t.Errorf("can't execute query: %v", err)
	}

	 result, err := LoginForManagers("", "", db)
	if err != nil {
		t.Errorf("can't execute Login: %v", err)
	}

	if result != false {
		t.Error("Login result not false for empty table")
	}
}

func TestLoginManager_LoginOk(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()


	_, err = db.Exec(`
  CREATE TABLE managers (
   id INTEGER PRIMARY KEY AUTOINCREMENT,
  login TEXT NOT NULL UNIQUE,
  password TEXT NOT NULL)`)
	if err != nil {
		t.Errorf("can't execute Login: %v", err)
	}

	_, err = db.Exec(`insert into managers (login,password) values ("vasya", "secret")`)
	if err != nil {
		t.Errorf("can't execute Login: %v", err)
	}

	result, err := LoginForManagers("vasya", "secret" , db)
	if err != nil {
		t.Errorf("can't execute Login: %v", err)
	}

	if result != true {
		t.Error("Login result not true for existing account")
	}
}

func TestLoginManager_LoginNotOkForInvalidPassword(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Errorf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("can't close db: %v", err)
		}
	}()


	_, err = db.Exec(`
 CREATE TABLE managers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
 login TEXT NOT NULL UNIQUE,
 password TEXT NOT NULL)`)
	if err != nil {
		t.Errorf("can't execute Login: %v", err)
	}

	_, err = db.Exec(`INSERT INTO managers(id, login, password) VALUES (1, 'vasya', 'secret')`)
	if err != nil {
		t.Errorf("can't execute Login: %v", err)
	}

	_, err = LoginForManagers("vasya", "password", db)
	if !errors.Is(err, ErrInvalidPass) {
		t.Errorf("Not ErrInvalidPass error for invalid pass: %v", err)
	}
}
