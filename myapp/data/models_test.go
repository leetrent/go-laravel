package data

import (
	"fmt"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	db2 "github.com/upper/db/v4"
)

func TestNew(t *testing.T) {
	fakeDB, _, _ := sqlmock.New()
	defer fakeDB.Close()

	/////////////////////////////////////////////
	// POSTGRES
	/////////////////////////////////////////////
	err := os.Setenv("DATABASE_TYPE", "postgres")
	if err != nil {
		fmt.Println(err)
		return
	}

	m := New(fakeDB)
	if fmt.Sprintf("%T", m) != "data.Models" {
		t.Error("Wrong type", fmt.Sprintf("%T", m))
	}

	/////////////////////////////////////////////
	// MYSQL
	/////////////////////////////////////////////
	err = os.Setenv("DATABASE_TYPE", "mysql")
	if err != nil {
		fmt.Println(err)
		return
	}

	m = New(fakeDB)
	if fmt.Sprintf("%T", m) != "data.Models" {
		t.Error("Wrong type", fmt.Sprintf("%T", m))
	}
}

func TestGetInsertID(t *testing.T) {
	var id db2.ID

	/////////////////////////////////////////////
	// POSTGRES
	/////////////////////////////////////////////
	id = int64(1)
	returnedID := getInsertID(id)
	if fmt.Sprintf("%T", returnedID) != "int" {
		t.Error("wrong type returned")
	}

	/////////////////////////////////////////////
	// MYSQL
	/////////////////////////////////////////////
	id = 1
	returnedID = getInsertID(id)
	if fmt.Sprintf("%T", returnedID) != "int" {
		t.Error("wrong type returned")
	}
}
