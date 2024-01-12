package main

import (
	"database/sql"
	"github.com/11marek/cacheuserlibrary"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"strconv"
)

// UserDBHandler handles communication with the SQLite database.
type UserDBHandler struct {
	db *sql.DB
}

// NewUserDBHandler initializes a new SQLite database handler.
func NewUserDBHandler(dbPath string) (*UserDBHandler, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create the users table in the database if it does not exist.
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT
    )`)
	if err != nil {
		return nil, err
	}

	return &UserDBHandler{db: db}, nil
}

// InsertUser adds a new user to the database.
func (h *UserDBHandler) InsertUser() (int64, error) {
	result, err := h.db.Exec("INSERT INTO users DEFAULT VALUES")
	if err != nil {
		return 0, err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

// GetUser returns information about the user from the database.
func (h *UserDBHandler) GetUser(userID int64) (bool, error) {
	var exists bool
	var ii int32
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", userID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking user existence: %v\n", err)
	} else {
		log.Printf("User %d existence checked successfully (Count: %d)\n", userID, ii)
	}
	return exists, err
}

// Function to fill the database with sample data if records do not already exist.
func fillDatabaseWithSampleData(dbHandler *UserDBHandler, count int) {
	// Check if records already exist in the database.
	recordsExist, err := checkIfUserExist(dbHandler, count)
	if err != nil {
		log.Fatal(err)
	}

	if !recordsExist {
		// Records do not exist, so add them to the database.
		for i := 0; i < count; i++ {
			_, err := dbHandler.InsertUser()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// Function to check if records already exist in the database.
func checkIfUserExist(dbHandler *UserDBHandler, count int) (bool, error) {
	var exists bool

	// Check if at least one record exists.
	err := dbHandler.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users LIMIT 1)").Scan(&exists)
	if err != nil {
		log.Printf("Error checking if records exist: %v\n", err)
	}

	return exists, err
}

func main() {
	dbHandler, err := NewUserDBHandler("users.db")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the caching handler.
	cacheHandler := cacheuserlibrary.NewUserCacheHandler(1000, &cacheuserlibrary.DatabaseConfig{})

	// Initialize the HTTP router.
	router := gin.Default()

	router.GET("/user/:id", func(c *gin.Context) {
		userID := c.Param("id")

		// Check if the user exists in the cache.
		if cacheHandler.IsUserCached(userID) {
			c.JSON(http.StatusOK, gin.H{
				"userID": userID,
				"from":   "cache",
			})
			return
		}

		// Check if the user exists in the database.
		intUserID, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}

		exists, err := dbHandler.GetUser(intUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Add the user to the cache.
		cacheHandler.HandleCaching(userID)

		c.JSON(http.StatusOK, gin.H{
			"userID": userID,
			"from":   "database",
		})
	})

	// Fill the database with sample data (100 users).
	fillDatabaseWithSampleData(dbHandler, 100)

	// Run the HTTP server.
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
