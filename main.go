package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"golang.org/x/crypto/bcrypt"
)

// Task represents a todo task
type Task struct {
	ID         int
	Title      string
	DueDate    string
	CategoryID int
	IsDone     bool
	UserID     int
}

// User represents a user account
type User struct {
	ID            int
	Email         string
	HashedPassword string
	Name          string
}

// Category represents a task category
type Category struct {
	ID     int
	Title  string
	Color  string
	UserID int
}

// Storage variables
var (
	userStorage       []User
	authenticatedUser *User
	categoryStorage   []Category
	taskStorage       []Task
	serializationMode string
)

// Constants
const (
	userStoragePath               = "users.txt"
	JsonSerializationMode         = "json"
	MandarAvadriSerializationMode = "mandaravadri"
)

// Print formats user information
func (u User) Print() {
	fmt.Printf("User: ID=%d, Name=%s, Email=%s\n", u.ID, u.Name, u.Email)
}

func main() {
	fmt.Println("Welcome to Todo App")

	// Load initial user data
	if err := loadUserStorageFromFile(); err != nil {
		fmt.Printf("Error loading users: %v\n", err)
	}

	// Parse command-line flags
	serializeMode := flag.String("serialize-mode", MandarAvadriSerializationMode, "serialization mode (json or mandaravadri)")
	command := flag.String("command", "", "initial command to run")
	flag.Parse()

	// Set serialization mode
	switch *serializeMode {
	case MandarAvadriSerializationMode:
		serializationMode = MandarAvadriSerializationMode
	case JsonSerializationMode:
		serializationMode = JsonSerializationMode
	default:
		fmt.Println("Invalid serialization mode, defaulting to JSON")
		serializationMode = JsonSerializationMode
	}

	// Main command loop
	scanner := bufio.NewScanner(os.Stdin)
	currentCommand := *command

	for {
		if currentCommand == "" {
			fmt.Print("Enter command (register-user, login, create-task, create-category, list-task, exit): ")
			scanner.Scan()
			currentCommand = scanner.Text()
		}

		if err := runCommand(currentCommand); err != nil {
			fmt.Printf("Error executing command: %v\n", err)
		}

		currentCommand = ""
	}
}

func runCommand(command string) error {
	// Require authentication for all commands except register-user and exit
	if command != "register-user" && command != "exit" && authenticatedUser == nil {
		if err := login(); err != nil {
			return fmt.Errorf("login failed: %v", err)
		}
	}

	switch command {
	case "create-task":
		return createTask()
	case "create-category":
		return createCategory()
	case "register-user":
		return registerUser()
	case "list-task":
		return listTasks()
	case "login":
		return login()
	case "exit":
		fmt.Println("Exiting Todo App")
		os.Exit(0)
	default:
		return fmt.Errorf("invalid command: %s", command)
	}
	return nil
}

func createTask() error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter task title: ")
	scanner.Scan()
	title := strings.TrimSpace(scanner.Text())
	if title == "" {
		return errors.New("task title cannot be empty")
	}

	fmt.Print("Enter category ID: ")
	scanner.Scan()
	categoryIDStr := strings.TrimSpace(scanner.Text())
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		return fmt.Errorf("invalid category ID: %v", err)
	}

	// Verify category exists and belongs to user
	isFound := false
	for _, c := range categoryStorage {
		if c.ID == categoryID && c.UserID == authenticatedUser.ID {
			isFound = true
			break
		}
	}
	if !isFound {
		return fmt.Errorf("category ID %d not found or doesn't belong to user", categoryID)
	}

	fmt.Print("Enter due date (YYYY-MM-DD): ")
	scanner.Scan()
	dueDate := strings.TrimSpace(scanner.Text())
	if dueDate == "" {
		return errors.New("due date cannot be empty")
	}

	task := Task{
		ID:         len(taskStorage) + 1,
		Title:      title,
		DueDate:    dueDate,
		CategoryID: categoryID,
		IsDone:     false,
		UserID:     authenticatedUser.ID,
	}

	taskStorage = append(taskStorage, task)
	fmt.Println("Task created successfully")
	return saveTasksToFile()
}

func createCategory() error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter category title: ")
	scanner.Scan()
	title := strings.TrimSpace(scanner.Text())
	if title == "" {
		return errors.New("category title cannot be empty")
	}

	fmt.Print("Enter category color: ")
	scanner.Scan()
	color := strings.TrimSpace(scanner.Text())
	if color == "" {
		return errors.New("category color cannot be empty")
	}

	category := Category{
		ID:     len(categoryStorage) + 1,
		Title:  title,
		Color:  color,
		UserID: authenticatedUser.ID,
	}

	categoryStorage = append(categoryStorage, category)
	fmt.Printf("Category created: %s (%s)\n", title, color)
	return saveCategoriesToFile()
}

func registerUser() error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter your name: ")
	scanner.Scan()
	name := strings.TrimSpace(scanner.Text())
	if name == "" {
		return errors.New("name cannot be empty")
	}

	fmt.Print("Enter your email: ")
	scanner.Scan()
	email := strings.TrimSpace(scanner.Text())
	if email == "" {
		return errors.New("email cannot be empty")
	}

	// Check if email already exists
	for _, u := range userStorage {
		if u.Email == email {
			return fmt.Errorf("email %s already registered", email)
		}
	}

	fmt.Print("Enter your password: ")
	scanner.Scan()
	password := strings.TrimSpace(scanner.Text())
	if password == "" {
		return errors.New("password cannot be empty")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	user := User{
		ID:            len(userStorage) + 1,
		Email:         email,
		HashedPassword: string(hashedPassword),
		Name:          name,
	}

	userStorage = append(userStorage, user)
	fmt.Printf("User registered: %s (%s)\n", name, email)
	return writeUserToFile(user)
}

func login() error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter your email: ")
	scanner.Scan()
	email := strings.TrimSpace(scanner.Text())

	fmt.Print("Enter your password: ")
	scanner.Scan()
	password := strings.TrimSpace(scanner.Text())

	for _, user := range userStorage {
		if user.Email == email {
			if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err == nil {
				authenticatedUser = &user
				fmt.Printf("Login successful! Welcome %s\n", user.Name)
				return nil
			}
		}
	}

	return errors.New("invalid email or password")
}

func listTasks() error {
	if len(taskStorage) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	fmt.Println("\nYour Tasks:")
	for _, task := range taskStorage {
		if task.UserID == authenticatedUser.ID {
			// Find category name
			categoryName := "Unknown"
			for _, cat := range categoryStorage {
				if cat.ID == task.CategoryID {
					categoryName = cat.Title
					break
				}
			}
			fmt.Printf("ID: %d, Title: %s, Category: %s, Due: %s, Done: %v\n",
				task.ID, task.Title, categoryName, task.DueDate, task.IsDone)
		}
	}
	return nil
}

func loadUserStorageFromFile() error {
	file, err := os.Open(userStoragePath)
	if os.IsNotExist(err) {
		return nil // File doesn't exist yet, which is fine
	}
	if err != nil {
		return fmt.Errorf("cannot open users file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		userStr := strings.TrimSpace(scanner.Text())
		if userStr == "" {
			continue
		}

		user, err := deserializeFromMandaravadri(userStr)
		if err != nil {
			fmt.Printf("Error deserializing user: %v\n", err)
			continue
		}
		userStorage = append(userStorage, user)
	}

	return scanner.Err()
}

func writeUserToFile(user User) error {
	file, err := os.OpenFile(userStoragePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot open users file: %v", err)
	}
	defer file.Close()

	var data []byte
	if serializationMode == MandarAvadriSerializationMode {
		data = []byte(fmt.Sprintf("id: %d, name: %s, email: %s, hashed_password: %s\n",
			user.ID, user.Name, user.Email, user.HashedPassword))
	} else {
		data, err = json.Marshal(user)
		if err != nil {
			return fmt.Errorf("cannot marshal user to JSON: %v", err)
		}
		data = append(data, '\n')
	}

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("cannot write to users file: %v", err)
	}
	return nil
}

func saveTasksToFile() error {
	file, err := os.OpenFile("tasks.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("cannot open tasks file: %v", err)
	}
	defer file.Close()

	for _, task := range taskStorage {
		data, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("cannot marshal task to JSON: %v", err)
		}
		if _, err := file.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("cannot write to tasks file: %v", err)
		}
	}
	return nil
}

func saveCategoriesToFile() error {
	file, err := os.OpenFile("categories.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("cannot open categories file: %v", err)
	}
	defer file.Close()

	for _, category := range categoryStorage {
		data, err := json.Marshal(category)
		if err != nil {
			return fmt.Errorf("cannot marshal category to JSON: %v", err)
		}
		if _, err := file.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("cannot write to categories file: %v", err)
		}
	}
	return nil
}

func deserializeFromMandaravadri(userStr string) (User, error) {
	if userStr == "" {
		return User{}, errors.New("empty user string")
	}

	user := User{}
	fields := strings.Split(userStr, ",")
	for _, field := range fields {
		parts := strings.SplitN(strings.TrimSpace(field), ": ", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]
		switch key {
		case "id":
			id, err := strconv.Atoi(value)
			if err != nil {
				return User{}, fmt.Errorf("invalid id: %v", err)
			}
			user.ID = id
		case "name":
			user.Name = value
		case "email":
			user.Email = value
		case "hashed_password":
			user.HashedPassword = value
		}
	}

	if user.Email == "" || user.Name == "" || user.HashedPassword == "" {
		return User{}, errors.New("incomplete user data")
	}
	return user, nil
}