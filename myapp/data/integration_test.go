//go:build integration

// run tests with this command: go test . --tags integration --count=1
// go test -cover . --tags integration

package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	host     = "localhost"
	user     = "postgres"
	password = "secret"
	dbName   = "celeritas_test"
	port     = "5435"
	dsn      = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var dummyUser = User{
	FirstName: "Some",
	LastName:  "Guy",
	Email:     "me@here.com",
	Active:    1,
	Password:  "password",
}

var models Models
var testDB *sql.DB
var resource *dockertest.Resource
var pool *dockertest.Pool

func TestMain(m *testing.M) {
	os.Setenv("DATABASE_TYPE", "postgres")

	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("ERROR(dockertest.NewPool): %s", err)
	}

	pool = p

	log.Println("pool....:", pool)
	log.Println("p.......:", p)
	log.Println("user....:", user)
	log.Println("password:", password)
	log.Println("dbName..:", dbName)
	log.Println("port....:", port)

	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13.4",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}
	log.Println("opts:", opts)

	resource, err = pool.RunWithOptions(&opts)
	log.Println("resource:", resource)
	if err != nil {
		if resource != nil {
			_ = pool.Purge(resource)
		}
		log.Fatalf("ERROR(pool.RunWithOptions): %s", err)
	}

	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbName))
		if err != nil {
			log.Println("ERROR(sql.Open):", err)
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("Error(pool.Retry): %s", err)
	}

	err = createTables(testDB)
	if err != nil {
		log.Println("ERROR(createTables):", err)
	}

	models = New(testDB)

	code := m.Run()

	// if err := pool.Purge(resource); err != nil {
	// 	log.Fatalf("could not purge resource: %s", err)
	// }

	os.Exit(code)
}

func createTables(db *sql.DB) error {
	log.Println("[createTable] => (BEGIN)")

	stmt := `
       CREATE OR REPLACE FUNCTION trigger_set_timestamp()
    RETURNS TRIGGER AS $$
    BEGIN
      NEW.updated_at = NOW();
      RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;
     
    drop table if exists users cascade;
     
    CREATE TABLE users (
        id SERIAL PRIMARY KEY,
        first_name character varying(255) NOT NULL,
        last_name character varying(255) NOT NULL,
        user_active integer NOT NULL DEFAULT 0,
        email character varying(255) NOT NULL UNIQUE,
        password character varying(60) NOT NULL,
        created_at timestamp without time zone NOT NULL DEFAULT now(),
        updated_at timestamp without time zone NOT NULL DEFAULT now()
    );
     
    CREATE TRIGGER set_timestamp
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_set_timestamp();
     
    drop table if exists remember_tokens;
     
    CREATE TABLE remember_tokens (
        id SERIAL PRIMARY KEY,
        user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
        remember_token character varying(100) NOT NULL,
        created_at timestamp without time zone NOT NULL DEFAULT now(),
        updated_at timestamp without time zone NOT NULL DEFAULT now()
    );
     
    CREATE TRIGGER set_timestamp
    BEFORE UPDATE ON remember_tokens
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_set_timestamp();
     
    drop table if exists tokens;
     
    CREATE TABLE tokens (
        id SERIAL PRIMARY KEY,
        user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
        first_name character varying(255) NOT NULL,
        email character varying(255) NOT NULL,
        token character varying(255) NOT NULL,
        token_hash bytea NOT NULL,
        created_at timestamp without time zone NOT NULL DEFAULT now(),
        updated_at timestamp without time zone NOT NULL DEFAULT now(),
        expiry timestamp without time zone NOT NULL
    );
     
    CREATE TRIGGER set_timestamp
    BEFORE UPDATE ON tokens
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_set_timestamp();
       `

	_, err := db.Exec(stmt)
	if err != nil {
		return err
	}

	log.Println("[createTable] => (END, returning nil)")
	return nil
}

/////////////////////////////////////////////////////////////////////
// TEST USERS TABLE FUNCTIONALITY
/////////////////////////////////////////////////////////////////////

func TestUser_Table(t *testing.T) {
	s := models.Users.Table()
	if s != "users" {
		t.Error("wrong table name returned (expected 'users'): ", s)
	}
}

func TestUser_Insert(t *testing.T) {
	id, err := models.Users.Insert(dummyUser)
	if err != nil {
		t.Error("failed to insert user: ", err)
	}

	if id == 0 {
		t.Error("0 returned as id after insert")
	}
}

func TestUser_Get(t *testing.T) {
	u, err := models.Users.GetByID(1)
	if err != nil {
		t.Error("failed to get user: ", err)
	}

	if u.ID == 0 {
		t.Error("id of returned user is 0: ", err)
	}
}

func TestUser_GetAll(t *testing.T) {
	_, err := models.Users.GetAll()
	if err != nil {
		t.Error("failed to get user: ", err)
	}
}

func TestUser_GetByEmail(t *testing.T) {
	u, err := models.Users.GetByEmail("me@here.com")
	if err != nil {
		t.Error("failed to get user: ", err)
	}

	if u.ID == 0 {
		t.Error("id of returned user is 0: ", err)
	}
}

func TestUser_Update(t *testing.T) {
	u, err := models.Users.GetByID(1)
	if err != nil {
		t.Error("failed to get user: ", err)
	}

	u.LastName = "Smith"
	err = u.Update(*u)
	if err != nil {
		t.Error("failed to update user: ", err)
	}

	u, err = models.Users.GetByID(1)
	if err != nil {
		t.Error("failed to get user: ", err)
	}

	if u.LastName != "Smith" {
		t.Error("last name not updated in database")
	}
}

func TestUser_PasswordMatches(t *testing.T) {
	u, err := models.Users.GetByID(1)
	if err != nil {
		t.Error("[password matches] => user with an ID of 1 was not found: ", err)
	}

	matches, err := u.PasswordMatches("password")
	if err != nil {
		t.Error("[password matches] => error encountered when calling user.PasswordMatches: ", err)
	}

	if !matches {
		t.Error("password doesn't match when it should.")
	}

	matches, err = u.PasswordMatches("123")
	if err != nil {
		t.Error("[password matches] => error encountered when calling user.PasswordMatches: ", err)
	}

	if !matches {
		t.Error("password matches when it should not.")
	}
}

func TestUser_ResetPassword(t *testing.T) {
	err := models.Users.ResetPassword(1, "new_password")
	if err != nil {
		t.Error("error resetting password: ", err)
	}

	err = models.Users.ResetPassword(2, "new_password")
	if err == nil {
		t.Error("did not get an error when trying to reset password for non-existent user")
	}
}

func TestUser_Delete(t *testing.T) {
	err := models.Users.Delete(1)
	if err != nil {
		t.Error("failed to delete user: " , err)
	}

	_, err = models.Users.GetByID(1)
	if err == nil {
		t.Error("retrieved user who was supposed to be deleted")
	}
}

/////////////////////////////////////////////////////////////////////
// TEST USERS TABLE FUNCTIONALITY
/////////////////////////////////////////////////////////////////////

// func TestToken_Table(t *testing.T) {
// 	s := models.Tokens.Table()
// 	if s != "tokens" {
// 		t.Error("wrong table returned (expected 'tokens')")
// 	}
// }

// func TestToken_GenerateToken(t *testing.T) {
// 	id, err := models.Users.Insert(dummyUser)
// 	if err != nil {
// 		t.Error("error inserting user: ", err)
// 	}
// }
