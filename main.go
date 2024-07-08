package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/novalyezu/simplebank-backend/api"
	db "github.com/novalyezu/simplebank-backend/db/sqlc"
	"github.com/novalyezu/simplebank-backend/token"
)

func main() {
	errEnv := godotenv.Load(".env")
	if errEnv != nil {
		log.Fatal("Error loading .env file")
	}

	var (
		dbHost            = os.Getenv("POSTGRES_HOST")
		dbPort            = os.Getenv("POSTGRES_PORT")
		dbUser            = os.Getenv("POSTGRES_USER")
		dbPass            = os.Getenv("POSTGRES_PASSWORD")
		dbName            = os.Getenv("POSTGRES_DATABASE")
		tokenSymmetricKey = os.Getenv("TOKEN_SYMMETRIC_KEY")
	)

	dbDriver := "postgres"
	dbSource := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("Cannot connect to db: ", err)
	}

	store := db.NewStore(conn)
	tokenMaker, err := token.NewPasetoMaker(tokenSymmetricKey)
	if err != nil {
		log.Fatal("Cannot create token maker: ", err)
	}

	server := api.NewServer(store, tokenMaker)

	err = server.Start(":3000")
	if err != nil {
		log.Fatal("Cannot start the server: ", err)
	}
}
