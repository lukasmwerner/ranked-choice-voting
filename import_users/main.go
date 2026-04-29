package importusers

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/osu-acm/acm-votes/database"
)

func Main() {

	db, err := sql.Open("sqlite3", "data/data.db")
	defer db.Close()
	if err != nil {
		log.Println("error opening database", err.Error())
		return
	}

	if len(os.Args) < 3 {
		log.Fatalln("no user csv file passed")
		return
	}
	userFileList := os.Args[2]

	membersCsv, err := os.Open(userFileList)
	if err != nil {
		return
	}
	csv_reader := csv.NewReader(membersCsv)
	_, err = csv_reader.Read()

	queries := database.New(db)

	i := 0
	for {
		row, err := csv_reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("error in reading csv:", err.Error())
			return
		}
		i += 1
		err = queries.AddUser(context.Background(), database.AddUserParams{
			Name:  row[1],
			Email: row[4],
		})
		if err != nil {
			fmt.Printf("error %d saving to database: %s\n", i, err.Error())
		}
		fmt.Printf("%d\r", i)
	}
	fmt.Printf("done with: %d records \n", i)

}
