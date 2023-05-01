package dbrepo

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
	"web-app/pkg/data"
	"web-app/pkg/repository"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	host     = "localhost"
	user     = "postgres"
	password = "postgres"
	dbName   = "users_test"
	port     = "5435"
	dsn      = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var resource *dockertest.Resource
var pool *dockertest.Pool
var testDB *sql.DB
var testRepo repository.DatabaseRepo

func TestMain(m *testing.M) {
	// connect to docker; fail if docker not running
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker; is it running? %s", err)
	}

	pool = p

	// set up our docker options, specifying the image and so forth
	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14.5",
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

	// get a resource (docker image)
	resource, err = pool.RunWithOptions(&opts)
	if err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not start resource: %s", err)
	}

	// start the image and wait until it's ready
	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbName))
		if err != nil {
			log.Println("error:", err)
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not connect to database: %s", err)
	}

	// populate the database with empty tables
	err = createTables()
	if err != nil {
		log.Fatalf("error creating tables: %s", err)
	}

	testRepo = &PostgresDBRepo{DB: testDB}

	// run tests
	code := m.Run()

	// clean up
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}

	os.Exit(code)
}

func createTables() error {
	tableSQL, err := os.ReadFile("./testdata/users.sql")
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = testDB.Exec(string(tableSQL))
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func Test_pingDB(t *testing.T) {
	err := testDB.Ping()
	if err != nil {
		t.Error("can't ping database")
	}
}

func TestPostgresDBRepoInsertUser(t *testing.T) {
	testUser := data.User{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
		Password:  "secret",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := testRepo.InsertUser(testUser)
	if err != nil {
		t.Errorf("insert user returned an error: %s", err)
	}
	if id != 1 {
		t.Errorf("insert user returned wrong ID. expected 1, but got %d", id)
	}
}

func TestPostgresDBRepoAllUsers(t *testing.T) {
	users, err := testRepo.AllUsers()
	if err != nil {
		t.Errorf("all users returned an error: %s", err)
	}
	if len(users) != 1 {
		t.Errorf("all users returned wrong size. expected 1, but got %d", len(users))
	}
	testUser := data.User{
		FirstName: "Jack",
		LastName:  "Smith",
		Email:     "jacksmith@example.com",
		Password:  "secret",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, _ = testRepo.InsertUser(testUser)
	users, err = testRepo.AllUsers()
	if err != nil {
		t.Errorf("all users returned an error: %s", err)
	}
	if len(users) != 2 {
		t.Errorf("all users returned wrong size after insert. expected 2, but got %d", len(users))
	}
}

func TestPostgresDBRepoGetUser(t *testing.T) {
	user, err := testRepo.GetUser(1)
	if err != nil {
		t.Errorf("error getting user by id: %s", err)
	}
	if user.Email != "admin@example.com" {
		t.Errorf("wrong email returned by getUser. expected admin@example.com, but got %s", user.Email)
	}
	_, err = testRepo.GetUser(3)
	if err == nil {
		t.Error("expected error from getting non existant user by id, but didn't get one")
	}
}

func TestPostgresDBRepoGetUserByEmail(t *testing.T) {
	user, err := testRepo.GetUserByEmail("jacksmith@example.com")
	if err != nil {
		t.Errorf("error getting user by email: %s", err)
	}
	if user.ID != 2 {
		t.Errorf("wrong id returned by getUserByEmail. expected 2, but got %d", user.ID)
	}
	_, err = testRepo.GetUserByEmail("bad@email.com")
	if err == nil {
		t.Error("expected error from getting user by email, but didn't get one")
	}
}

func TestPostgresDBRepoUpdateUser(t *testing.T) {
	user, _ := testRepo.GetUser(1)
	user.FirstName = "James"
	user.Email = "jamessmith@example.com"
	err := testRepo.UpdateUser(*user)
	if err != nil {
		t.Errorf("update user returned an error: %s", err)
	}
	user, _ = testRepo.GetUser(1)
	if user.FirstName != "James" {
		t.Errorf("expected user first name to be James, but got %s", user.FirstName)
	}
	if user.Email != "jamessmith@example.com" {
		t.Errorf("expected user email to be jamessmith@example.com, but got %s", user.Email)
	}
}

func TestPostgresDBRepoDeleteUser(t *testing.T) {
	err := testRepo.DeleteUser(2)
	if err != nil {
		t.Errorf("error deleting user with id 2: %s", err)
	}
	_, err = testRepo.GetUser(2)
	if err == nil {
		t.Error("user with id 2 should have been deleted, but wasn't")
	}
}

func TestPostgresDBRepoResetPassword(t *testing.T) {
	newPassword := "newpassword"
	err := testRepo.ResetPassword(1, newPassword)
	if err != nil {
		t.Errorf("reset password returned an error: %s", err)
	}
	user, _ := testRepo.GetUser(1)
	matches, err := user.PasswordMatches(newPassword)
	if err != nil {
		t.Error(err)
	}
	if !matches {
		t.Error("reset password: password should match, but does not")
	}
}

func TestPostgresDBRepoInsertUserImage(t *testing.T) {
	userImage := data.UserImage{
		UserID:    "1",
		FileName:  "test.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	id, err := testRepo.InsertUserImage(userImage)
	if err != nil {
		t.Errorf("insert user image returned an unexpected error: %s", err)
	}
	if id != 1 {
		t.Errorf("insert user image: expected id to be 1, but got %d", id)
	}
	userImage = data.UserImage{
		UserID:    "50",
		FileName:  "test.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = testRepo.InsertUserImage(userImage)
	if err == nil {
		t.Error("insert user image: expected an error due to incorrect userID, but didn't get one")
	}
}
