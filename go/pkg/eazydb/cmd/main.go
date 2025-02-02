package main

import (
	"github.com/mattam808/eazydb/go/pkg/eazydb"
	"github.com/mattam808/eazydb/go/pkg/eazydb/dbtypes"
	"github.com/sirupsen/logrus"
)

type Users struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type User struct {
	ID          int    `json:"id" faker:"-"`
	Name        string `json:"name" faker:"name"`
	Email       string `json:"email" faker:"email"`
	Title       string `json:"title" faker:"name"`
	Age         int    `json:"age" faker:"boundary_start=18, boundary_end=60"`
	DaysPresent int    `json:"days_present" faker:"boundary_start=18, boundary_end=60"`
}

func main() {
	log := logrus.New()

	c, err := eazydb.NewClient(eazydb.ClientOptions{
		User:     "postgres",
		Password: "postgres",
		Host:     "localhost",
		Port:     "5432",
		Name:     "postgres",
		Type:     eazydb.POSTGRES,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a table
	_, err = c.NewTable("users").Fields(User{}).Key("id", dbtypes.SERIAL).Exec()
	if err != nil {
		log.Fatalf("could not create table %v", err)
	}

	// Insert data
	users := makeUsers(5000)
	metadata, err := c.Table("users").Add(users).Exec()
	if err != nil {
		log.Fatalf("could not insert users table %v", err)
	}

	log.Infof("Users created: %v, query time: %v", metadata.RowsAffected, metadata.Duration)

	// Update a field
	Mat := &User{
		Age: 24,
	}
	metadata, err = c.Table("users").Update(Mat).Where(
		*eazydb.String("name").Equals("Mat"),
	).Exec()

	if err != nil {
		log.Fatalf("could not update users table %v", err)
	}

	log.Infof("Users updated: %v, query time: %v", metadata.RowsAffected, metadata.Duration)

	// Delete
	metadata, err = c.Table("users").Delete().Where(
		*eazydb.Int("age").Equals(40),
	).Exec()
	if err != nil {
		log.Fatalf("could not delete from users table %v", err)
	}
	log.Infof("Users updated: %v, query time: %v", metadata.RowsAffected, metadata.Duration)

	// Get fields, the best part is parsing directly to an object
	var resp []User
	_, err = c.Table("users").Get(User{}).Where(
		*eazydb.Int("age").Equals(24),
	).Exec(&resp)

	if err != nil {
		log.Fatalf("could not get from users table %v", err)
	}

	log.Infof("users returned %v", len(resp))
	dump("./output.json", resp)
}
