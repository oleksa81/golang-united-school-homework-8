package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

var (
	operationMissingErr = errors.New("-operation flag has to be specified")
	fileNameMissingErr  = errors.New("-fileName flag has to be specified")
	idMissingErr        = errors.New("-id flag has to be specified")
	itemMissingErr      = errors.New("-item flag has to be specified")
)

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type Arguments map[string]string

func Perform(args Arguments, writer io.Writer) error {
	if args["operation"] == "add" {
		err := AddNewItem(args, writer)
		if err != nil {
			return err
		}
	} else if args["operation"] == "list" {
		err := GetInfo(args, writer)
		if err != nil {
			return err
		}
	} else if args["operation"] == "findById" {
		err := FindByID(args, writer)
		if err != nil {
			return err
		}
	} else if args["operation"] == "remove" {
		err := RemoveUser(args, writer)
		if err != nil {
			return err
		}
	} else if len(args["operation"]) == 0 {
		return operationMissingErr
	} else {
		return fmt.Errorf("Operation %s not allowed!", args["operation"])
	}

	return nil
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

func parseArgs() Arguments {
	_ = os.Args[1:]
	id := flag.String("id", "", "chose user id")
	inputBody := flag.String("item", "", "body for add")
	choseOperation := flag.String("operation", "", "chose operation")
	choseFile := flag.String("fileName", "", "chose file")
	flag.Parse()
	mp := Arguments{}
	mp["id"] = *id
	mp["operation"] = *choseOperation
	mp["item"] = *inputBody
	mp["fileName"] = *choseFile
	return mp
}

func GetInfo(args Arguments, writer io.Writer) error {
	fileName := args["fileName"]
	if len(fileName) == 0 {
		return fileNameMissingErr
	}
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	writer.Write(data)
	return nil
}

func AddNewItem(args Arguments, writer io.Writer) error {
	fileName := args["fileName"]
	item := args["item"]
	if len(fileName) == 0 {
		return fileNameMissingErr
	}
	if len(item) == 0 {
		return itemMissingErr
	}
	input := User{}
	oldData := []User{}
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	err = json.Unmarshal([]byte(item), &input)
	if err != nil {
		return err
	}
	if check := IsValid(input); !check {
		return errors.New("invalid input")
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	if len(data) != 0 {
		err = json.Unmarshal(data, &oldData)
		if err != nil {
			return err
		}
	}

	for _, i := range oldData {
		if i.Id == input.Id {
			str := fmt.Sprintf("Item with id %s already exists", input.Id)
			writer.Write([]byte(str))
			return nil
			// return errors.New(str)
		}
	}
	oldData = append(oldData, input)
	out, err := json.Marshal(oldData)
	if err != nil {
		return err
	}
	err = os.WriteFile(fileName, out, 0644)
	if err != nil {
		return err
	}
	return nil
}

func RemoveUser(args Arguments, writer io.Writer) error {
	fileName := args["fileName"]
	id := args["id"]
	fileBody := []User{}
	check := false
	newData := []User{}
	if len(fileName) == 0 {
		return fileNameMissingErr
	}
	if len(id) == 0 {
		return idMissingErr
	}
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &fileBody)
	if err != nil {
		return err
	}
	for _, i := range fileBody {
		if i.Id == id {
			check = true
		} else if i.Id != id {
			newData = append(newData, i)
		}
	}
	if !check {
		str := fmt.Sprintf("Item with id %s not found", id)
		writer.Write([]byte(str))
		return nil
	}
	afterRemove, err := json.Marshal(newData)
	if err != nil {
		return err
	}

	err = os.WriteFile(fileName, afterRemove, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

func FindByID(args Arguments, writer io.Writer) error {
	input := []User{}
	id := args["id"]
	fileName := args["fileName"]
	if len(id) == 0 {
		return idMissingErr
	}
	if len(fileName) == 0 {
		return fileNameMissingErr
	}
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &input)
	if err != nil {
		return err
	}
	ind := 0
	check := false
	for index, i := range input {
		if id == i.Id {
			ind = index
			check = true
		}
	}
	if !check {
		writer.Write([]byte(""))
		return nil
	}
	out, err := json.Marshal(input[ind])
	if err != nil {
		return err
	}
	writer.Write(out)
	return nil
}

func IsValid(u User) bool {
	if u.Age == 0 {
		return false
	}
	if len(u.Email) == 0 {
		return false
	}
	if len(u.Id) == 0 {
		return false
	}
	return true
}
