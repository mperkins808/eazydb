package main

import (
	"encoding/json"
	"os"

	"github.com/bxcodec/faker/v3"
)

func makeUsers(amount int) []User {
	users := make([]User, amount)
	for i := 0; i < amount; i++ {
		faker.FakeData(&users[i]) // Populate the user data with fake values

		ages, _ := faker.RandomInt(16, 50)
		users[i].Age = ages[0]
	}
	return users
}

func dump(path string, data any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}
