package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	testQueries *Queries
)

func TestMain(m *testing.M) {
	errEnv := godotenv.Load("../../.env")
	if errEnv != nil {
		log.Fatal("Error loading .env file")
	}

	var (
		dbHost = os.Getenv("POSTGRES_HOST")
		dbPort = os.Getenv("POSTGRES_PORT")
		dbUser = os.Getenv("POSTGRES_USER")
		dbPass = os.Getenv("POSTGRES_PASSWORD")
		dbName = os.Getenv("POSTGRES_DATABASE")
	)

	dbDriver := "postgres"
	dbSource := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("Cannot connect to db: ", err)
	}

	testQueries = New(conn)

	os.Exit(m.Run())
}
