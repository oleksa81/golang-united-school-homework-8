package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"A
	"os"
)

type Arguments map[string]string

const (
	idFlag        = "id"
	itemFlag      = "item"
	operationFlag = "operation"
	filenameFlag  = "fileName"
)

const (
	opAdd      = "add"
	opList     = "list"
	opFindById = "findById"
	opRemove   = "remove"
)

var _ = flag.String(idFlag, "", "user id")
var _ = flag.String(itemFlag, "", "user json payload")
var _ = flag.String(operationFlag, "", "operation to be performed")
var _ = flag.String(filenameFlag, "", "path to json file")

func parseArgs() Arguments {
	flag.Parse()
	result := Arguments{}
	flag.Visit(func(flag *flag.Flag) {
		result[flag.Name] = flag.Value.String()
	})
	return result
}

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func readUsers(args Arguments) ([]User, error) {
	filename := args[filenameFlag]
	if filename == "" {
		return nil, fmt.Errorf("-fileName flag has to be specified")
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	var users []User
	err = json.Unmarshal(data, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func writeUsers(users []User, writer io.Writer) error {
	data, err := json.Marshal(users)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
	//return json.NewEncoder(writer).Encode(users)  - appends with '\n', tests do not match
}

func writeUsersToFile(args Arguments, users []User) error {
	filename := args[filenameFlag]
	if filename == "" {
		return fmt.Errorf("-fileName flag has to be specified")
	}
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)
	return writeUsers(users, file)
}

func doFindById(users []User, id string) int {
	for i, user := range users {
		if user.Id == id {
			return i
		}
	}
	return -1
}

func add(args Arguments, writer io.Writer) error {
	item := args[itemFlag]
	if item == "" {
		return fmt.Errorf("-item flag has to be specified")
	}
	var user User
	err := json.Unmarshal([]byte(item), &user)
	if err != nil {
		return err
	}
	users, err := readUsers(args)
	if err != nil {
		return err
	}
	if doFindById(users, user.Id) >= 0 {
		_, err = fmt.Fprintf(writer, "Item with id %s already exists", user.Id)
		if err != nil {
			panic(err)
		}
		return nil
	}
	users = append(users, user)
	return writeUsersToFile(args, users)
}

func list(args Arguments, writer io.Writer) error {
	users, err := readUsers(args)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return nil
	}
	return writeUsers(users, writer)
}

func findById(args Arguments, writer io.Writer) error {
	id := args[idFlag]
	if id == "" {
		return fmt.Errorf("-id flag has to be specified")
	}
	users, err := readUsers(args)
	if err != nil {
		return err
	}

	found := doFindById(users, id)
	if found < 0 {
		return nil
	}
	data, err := json.Marshal(users[found])
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}

func remove(args Arguments, writer io.Writer) error {
	id := args[idFlag]
	if id == "" {
		return fmt.Errorf("-id flag has to be specified")
	}
	users, err := readUsers(args)
	if err != nil {
		return err
	}

	found := doFindById(users, id)
	if found < 0 {
		_, err = fmt.Fprintf(writer, "Item with id %s not found", id)
		if err != nil {
			panic(err)
		}
		return nil
	}
	result := make([]User, 0, len(users)-1)
	for _, user := range users {
		if user.Id != id {
			result = append(result, user)
		}
	}
	return writeUsersToFile(args, result)
}

func Perform(args Arguments, writer io.Writer) error {
	op, ok := args[operationFlag]
	if !ok || op == "" {
		return fmt.Errorf("-operation flag has to be specified")
	}

	switch op {
	case opAdd:
		return add(args, writer)
	case opList:
		return list(args, writer)
	case opFindById:
		return findById(args, writer)
	case opRemove:
		return remove(args, writer)
	default:
		//goland:noinspection GoErrorStringFormat
		return fmt.Errorf("Operation %v not allowed!", op)
	}
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
