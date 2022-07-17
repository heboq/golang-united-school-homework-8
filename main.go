package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

type Arguments map[string]string

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

var (
	errorIdFlagIsMissing        = errors.New("-id flag has to be specified")
	errorItemFlagIsMissing      = errors.New("-item flag has to be specified")
	errorFileNameFlagIsMissing  = errors.New("-fileName flag has to be specified")
	errorOperationFlagIsMissing = errors.New("-operation flag has to be specified")
)

func Perform(args Arguments, writer io.Writer) error {
	if args["fileName"] == "" {
		return errorFileNameFlagIsMissing
	}

	file, err := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	users := []User{}

	if len(data) != 0 {
		if err := json.Unmarshal(data, &users); err != nil {
			return err
		}
	}

	switch args["operation"] {
	case "":
		return errorOperationFlagIsMissing

	case "list":
		if len(users) == 0 {
			return nil
		}
		list, err := json.Marshal(users)
		if err != nil {
			return err
		}
		writer.Write(list)

	case "add":
		if args["item"] == "" {
			return errorItemFlagIsMissing
		}

		var item User

		if err := json.Unmarshal([]byte(args["item"]), &item); err != nil {
			return err
		}

		for _, user := range users {
			if user.ID == item.ID {
				writer.Write([]byte("Item with id " + item.ID + " already exists"))
				return nil
			}
		}

		users = append(users, item)

		newUsers, err := json.Marshal(users)
		if err != nil {
			return err
		}
		if err := os.WriteFile(args["fileName"], newUsers, 0644); err != nil {
			return err
		}

	case "remove":
		if args["id"] == "" {
			return errorIdFlagIsMissing
		}

		for i, user := range users {
			if args["id"] == user.ID {
				users = append(users[:i], users[i+1:]...)
				newUsers, err := json.Marshal(users)
				if err != nil {
					return err
				}
				if err := os.WriteFile(args["fileName"], newUsers, 0644); err != nil {
					return err
				}
				return nil
			}
		}
		writer.Write([]byte(fmt.Sprintf("Item with id %s not found", args["id"])))

	case "findById":
		if args["id"] == "" {
			return errorIdFlagIsMissing
		}

		for _, user := range users {
			if args["id"] == user.ID {
				u, err := json.Marshal(user)
				if err != nil {
					return err
				}
				writer.Write(u)
				return nil
			}
		}
		writer.Write([]byte(""))
		return nil

	default:
		return fmt.Errorf("Operation %s not allowed!", args["operation"])
	}

	return nil
}

func parseArgs() Arguments {
	var (
		id        = flag.String("id", "", "user id")
		item      = flag.String("item", "", "valid json object with the id, email and age fields")
		fileName  = flag.String("fileName", "", "name of a JSON file")
		operation = flag.String("operation", "", "supported operations: add, list, findById, remove")
	)
	flag.Parse()
	return Arguments{
		"id":        *id,
		"item":      *item,
		"operation": *operation,
		"fileName":  *fileName,
	}
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
	fmt.Println()
}
