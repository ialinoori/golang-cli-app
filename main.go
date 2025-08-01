package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Task struct {
	ID         int
	Title      string
	DueDate    string
	CategoryId int
	IsDone     bool
	UserId     int
}

type User struct {
	ID       int
	Email    string
	Password string
	Name     string
}

type Category struct {
	ID     int
	Title  string
	Color  string
	UserId int
}

func (u User) print() {
	fmt.Println("user:", u.ID, u.Name, u.Email)
}

var userStorage []User
var authenticatedUser *User

var categoryStorage []Category
var taskStorage []Task

const userStoragePath = "user.txt"

func main() {
	fmt.Println("hello to Todo App")

	loadUserStorageFromFile()

	command := flag.String("command", "no command", "command to run")
	flag.Parse()

	for {
		runCommand(*command)

		scanner := bufio.NewScanner((os.Stdin))
		fmt.Println("pls enter another command")
		scanner.Scan()
		*command = scanner.Text()
	}

}

func runCommand(command string) {

	if command != "register-user" && command != "exit" && authenticatedUser == nil {
		login()

		if authenticatedUser == nil {
			return
		}
	}

	switch command {
	case "create-task":
		createTask()
	case "create-category":
		createCategory()
	case "register-user":
		registerUser()
	case "list-task":
		listTask()
	case "login":
		login()
	case "exit":
		os.Exit(0)
	default:
		fmt.Println("command is not valid")
	}
}

func createTask() {

	scanner := bufio.NewScanner(os.Stdin)

	var title, duedate, category string
	fmt.Println("pls enter the task title")
	scanner.Scan()
	title = scanner.Text()

	fmt.Println("pls enter the task category-id")
	scanner.Scan()
	category = scanner.Text()

	categoryId, err := strconv.Atoi(category)

	if err != nil {
		fmt.Println("not valid cate id")

		return
	}

	isFound := false
	for _, c := range categoryStorage {
		if c.ID == categoryId && c.UserId == authenticatedUser.ID {
			isFound = true
			break
		}
	}

	if !isFound {
		fmt.Println(" not found cate id")

		return
	}
	fmt.Println("pls enter the task duedate")
	scanner.Scan()
	duedate = scanner.Text()

	task := Task{
		ID:         len(taskStorage) + 1,
		Title:      title,
		DueDate:    duedate,
		CategoryId: categoryId,
		IsDone:     false,
		UserId:     authenticatedUser.ID,
	}
	taskStorage = append(taskStorage, task)

}
func createCategory() {
	scanner := bufio.NewScanner(os.Stdin)

	var title, color string

	fmt.Println("pls enter the category title")
	scanner.Scan()
	title = scanner.Text()

	fmt.Println("pls enter the category color")
	scanner.Scan()
	color = scanner.Text()

	fmt.Println("category:", title, color)

	category := Category{
		ID:     len(categoryStorage) + 1,
		Title:  title,
		Color:  color,
		UserId: authenticatedUser.ID,
	}

	categoryStorage = append(categoryStorage, category)
}
func registerUser() {
	scanner := bufio.NewScanner(os.Stdin)

	var id, name, email, password string

	fmt.Println("pls enter your anme")
	scanner.Scan()
	name = scanner.Text()

	fmt.Println("pls enter your email")
	scanner.Scan()
	email = scanner.Text()

	fmt.Println("pls enter your password")
	scanner.Scan()
	password = scanner.Text()

	id = email

	fmt.Println("user", email, password, id, name)

	user := User{
		ID:       len(userStorage) + 1,
		Email:    email,
		Password: password,
		Name:     name,
	}

	userStorage = append(userStorage, user)

	writeUserToFile(user)

}
func login() {
	scanner := bufio.NewScanner(os.Stdin)

	var email, password string

	fmt.Println("pls enter your tamil")
	scanner.Scan()
	email = scanner.Text()

	fmt.Println("pls enter your password")
	scanner.Scan()
	password = scanner.Text()

	for _, user := range userStorage {
		if user.Email == email && user.Password == password {
			authenticatedUser = &user

			break
		}
	}

	if authenticatedUser == nil {
		fmt.Println("the email or password is not correct")

	}

}

func listTask() {
	for _, task := range taskStorage {
		if task.UserId == authenticatedUser.ID {
			fmt.Println(task)

		}
	}
}

func loadUserStorageFromFile() {
	file, err := os.Open(userStoragePath)

	if err != nil {
		fmt.Println("cant open the file", err)
	}

	var data = make([]byte, 10240)

	_, oErr := file.Read(data)
	if oErr != nil {
		fmt.Println("cant read from the file", oErr)
	}

	var dataStr string = string(data)

	userSlice := strings.Split(dataStr, "\n")

	for _, u := range userSlice {
		if u == " " {
			continue
		}

		var user = User{}

		userFields := strings.Split(u, ",")

		for _, filed := range userFields {
			fmt.Println(filed)

			values := strings.Split(filed, ": ")

			if len(values) != 2 {
				fmt.Println("record is not valid")

				continue
			}

			fieldName := strings.ReplaceAll(values[0], " ", " ")
			fieldValue := values[1]

			switch fieldName {
			case "id":
				id, err := strconv.Atoi(fieldValue)

				if err != nil {
					fmt.Println("not valid")
				}
				user.ID = id
			case "name":
				user.Name = fieldValue
			case "email":
				user.Email = fieldValue
			case "password":
				user.Password = fieldValue

			}

		}
		fmt.Printf("user: %+v\n", user)

	}

}

func writeUserToFile(user User) {

	var file *os.File

	file, err := os.OpenFile(userStoragePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Println("can not create or open file", err)

		return
	}

	defer file.Close()

	data := fmt.Sprintf("id: %d, name: %s,email: %s,password:%s\n", user.ID, user.Name, user.Email, user.Password)

	var b = []byte(data)

	numberofWrittenBytes, wErr := file.Write(b)

	if wErr != nil {
		fmt.Println("can not write to the file", wErr)

		return
	}

	fmt.Println("numberofWrittenBytes", numberofWrittenBytes)

}
