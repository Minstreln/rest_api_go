package sqlconnect

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDb() (*sql.DB, error) {
	fmt.Println("Connecting to MariaDB...")
	// err := godotenv.Load()
	// if err != nil {
	// 	return nil, err
	// }

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	host := os.Getenv("HOST")

	// connectionString := "root:carlosdbserver234@tcp(127.0.0.1:3306)/" + os.Getenv("DB_NAME")
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbname)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		// panic(err)
		return nil, err
	}
	fmt.Println("Connected to MariaDB")
	return db, nil
}
