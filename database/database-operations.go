package database

import (
	"fmt"
	"os"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func InitializeDatabase(database *sql.DB, name string) {
	db = database

	fmt.Println("Checking for the existence of our tables")

	row := db.QueryRow("select count(*) from information_schema.tables where table_schema = ? and table_name = ?", name, "scaling_instances")

	var count int

	err := row.Scan(&count)
	if err != nil {
		fmt.Println("error in query")
		fmt.Println(err)
		os.Exit(1)
	} else {

		if count == 0 {
			fmt.Println("Tables not found, creating them...")
			_, err := db.Exec("CREATE table scaling_instances (instance_id varchar(64), name varchar(64), min_instances integer, max_instances integer, min_memory integer, max_memory integer, time_between_scales integer, time_over_threshold integer, current_state varchar(12) )")

			if err != nil {
				fmt.Println("error in create")
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			fmt.Println("tables already created, moving on")
		}
	}

}
