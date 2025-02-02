package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/bxcodec/faker/v3"
	"github.com/mattam808/eazydb/go/pkg/eazydb"
	"github.com/mattam808/eazydb/go/pkg/eazydb/dbtypes"
	"github.com/sirupsen/logrus"
)

type Users struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {

	log := logrus.New()

	c, err := eazydb.NewClient(eazydb.ClientOptions{
		User:       "postgres",
		Password:   "postgres",
		Host:       "localhost",
		Port:       "5432",
		Name:       "postgres",
		Type:       eazydb.POSTGRES,
		EnableLogs: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(test(c))

	_, err = c.NewTable("users").Fields(Users{}).Key("id", dbtypes.SERIAL).Exec()
	if err != nil {
		log.Fatal(err)
	}

	user := &Users{
		Name: "Mat",
		Age:  24,
	}

	m, err := c.Table("users").Add(user).Where(
		*eazydb.Int("age").Equals(23),
		*eazydb.Int("age").GreaterThan(10).Or(*eazydb.Int("age").NotEqual(4)),
	).Dry().Exec()
	if err != nil {
		log.Fatal(err)
	}
	log.Info(m.Query)

	var users []Users
	_, err = c.Table("users").Get(Users{}).Exec(&users)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(users)

}

type User struct {
	ID          int    `json:"id" faker:"-"`
	Name        string `json:"name" faker:"name"`
	Email       string `json:"email" faker:"email"`
	Title       string `json:"title" faker:"name"`
	Age         int    `json:"age" faker:"boundary_start=18, boundary_end=60"`
	DaysPresent int    `json:"days_present" faker:"boundary_start=18, boundary_end=60"`
}

func test(c *eazydb.Client) error {
	_, err := c.NewTable("users").Fields(User{}).Key("id", dbtypes.SERIAL).AddNewFields().Exec()
	if err != nil {
		return err
	}

	// gen users using faker
	// Generate 1000 fake users using faker
	users := make([]User, 1)
	for i := 0; i < 1; i++ {
		err := faker.FakeData(&users[i]) // Populate the user data with fake values
		if err != nil {
			return err
		}
		ages, err := faker.RandomInt(16, 50)
		if err != nil {
			return err
		}
		users[i].Age = ages[0]
	}

	for _, user := range users {
		_, err = c.Table("users").Add(user).Exec()
		if err != nil {
			return err
		}
	}

	update := &User{
		Email:       "test@example.com",
		Age:         45,
		Title:       "Boss",
		DaysPresent: 100,
	}

	_, err = c.Table("users").Update(update).Where(
		*eazydb.String("name").Equals("Mr. Don Franecki"),
	).Exec()
	if err != nil {
		return err
	}

	var output []User
	_, err = c.Table("users").Get(User{}).Where(
		*eazydb.String("name").Equals("Mr. Don Franecki"),
	).Exec(&output)
	if err != nil {
		return err
	}

	//write to local "output.json"
	// Write the fetched users to a local "output.json" file
	file, err := os.Create("output.json")
	if err != nil {
		return fmt.Errorf("failed to create output.json file: %v", err)
	}
	defer file.Close()

	// Marshal the output into JSON and write it to the file
	encoder := json.NewEncoder(file)
	err = encoder.Encode(output)
	if err != nil {
		return fmt.Errorf("failed to write users to output.json: %v", err)
	}

	log.Println("Successfully generated 1000 users and saved to output.json")
	return nil
}
