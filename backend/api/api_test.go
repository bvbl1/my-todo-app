package api

import (
	// Assuming 'db *sql.DB' is defined globally or accessible.
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os" // Needed for os.Exit and environment variables
	"testing"

	// Consider using the official gin-contrib/cors: "github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/cors" // Current: "github.com/gin-gonic/contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// router is the Gin engine used for tests.
var router *gin.Engine

// db is assumed to be your global database connection, initialized by SetupPostgres.
// Ensure this is properly declared and initialized in your actual package structure
// if it's not in this file. For example, in a models.go or db.go:
// var DB *sql.DB
// var db *sql.DB // Placeholder: replace with your actual global DB variable if named differently or in another file.

// ListItem struct definition.
// This should ideally be in a models.go or a shared types file.
// type ListItem struct {
// 	Id   string `json:"id"`
// 	Item string `json:"item"`
// 	Done bool   `json:"done"`
// }

/*
IMPORTANT: The following functions `SetupPostgres`, `TodoItems`, `CreateTodoItem`,
`UpdateTodoItem`, and `DeleteTodoItem` are assumed to be part of your 'api' package
or correctly imported. Their actual implementations are in your main application files.
The comments below highlight how they should behave for these tests to pass.
*/

// SetupPostgres (Assumed to be in your main code, e.g., db.go or main.go)
// CRITICAL: This function MUST use environment variables for DB connection
// parameters (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME) so that it
// can connect to the PostgreSQL service container in GitHub Actions.
// In GitHub Actions CI, DB_HOST should typically be "postgres" (the service name).
/*
func SetupPostgres() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	// ... get other env vars ...

	if dbHost == "" { dbHost = "localhost" } // Default for local if not set
	if dbPort == "" { dbPort = "5432" }    // Default for local if not set

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	var err error
	// 'db' should be your package-level or global database connection variable
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB connection: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}
	fmt.Println("Successfully connected to test database!")
	// Initialize schema if needed for tests, or ensure migrations run.
}
*/

// API Handlers (Assumed to be in your main code, e.g., handlers.go)
/*
func TodoItems(c *gin.Context) {
	// Must fetch items from 'db'.
	// If no items, MUST return: c.JSON(http.StatusOK, gin.H{"items": []ListItem{}}) // Empty slice, not nil
	// Otherwise: c.JSON(http.StatusOK, gin.H{"items": fetchedItems})
}

func CreateTodoItem(c *gin.Context) {
	itemName := c.Param("item")
	// Must save to 'db'. The database should assign an ID.
	// Must return the created item, including its new ID.
	// e.g., c.JSON(http.StatusCreated, gin.H{"item": createdItemWithID}) // Note: test expects key "items" for single create
}

func UpdateTodoItem(c *gin.Context) {
	// id := c.Param("id")
	// doneStatus := c.Param("done")
	// Must update item in 'db'.
	// If item not found, MUST return: c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
	// On success: c.JSON(http.StatusOK, gin.H{"item": updatedItem}) // Or just status OK
}

func DeleteTodoItem(c *gin.Context) {
	// id := c.Param("id")
	// Must delete item from 'db'.
	// If item not found, MUST return: c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
	// On success: c.JSON(http.StatusOK, gin.H{"message": "deleted"}) // Or just status OK
}
*/

// displayTable is a test helper.
func displayTable() {
	if db == nil {
		fmt.Println("displayTable: DB connection is nil. SetupPostgres might not have initialized 'db'.")
		return
	}
	rows, err := db.Query("SELECT id, item, done FROM list") // Be explicit with column names
	if err != nil {
		fmt.Println("displayTable query error:", err.Error())
		return
	}
	defer rows.Close()

	items := make([]ListItem, 0)
	for rows.Next() {
		item := ListItem{}
		if err := rows.Scan(&item.Id, &item.Item, &item.Done); err != nil {
			fmt.Println("displayTable scan error:", err.Error())
			// Optionally continue to allow partial display
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil { // Check for errors during iteration
		fmt.Println("displayTable rows iteration error:", err.Error())
	}
	fmt.Println("Current items in DB for test:", items)
}

// emptyTable is a test helper to clear the database state.
func emptyTable() {
	if db == nil {
		fmt.Println("emptyTable: DB connection is nil. Cannot clear table. Tests will be unreliable.")
		// This is a critical failure point if db is not initialized by SetupPostgres.
		return
	}
	// Using MustExec for simplicity in tests; errors would cause panic.
	// Or check errors explicitly:
	if _, err := db.Exec("DELETE FROM list;"); err != nil {
		fmt.Printf("emptyTable: Failed to delete from list: %v\n", err)
	}
	// Reset id counter. Ensure 'list_id_seq' is the correct sequence name for your table.
	if _, err := db.Exec("ALTER SEQUENCE list_id_seq RESTART WITH 1;"); err != nil {
		fmt.Printf("emptyTable: Failed to reset sequence list_id_seq: %v\n", err)
	}
}

// SetupRoutesForTest configures the Gin engine for testing.
// It's good practice to have this separate if your main SetupRoutes does more (e.g., global middleware).
func SetupRoutesForTest() *gin.Engine {
	r := gin.New()      // Use gin.New() for a clean engine in tests, add middleware selectively.
	r.Use(gin.Logger()) // Optional: logger for test runs
	r.Use(gin.Recovery())

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true // For testing convenience. Be stricter in production.
	r.Use(cors.New(config))

	// Routes should match your application's routes.
	// These handlers (TodoItems, etc.) are your actual application handlers.
	r.GET("/items", TodoItems)
	r.GET("/item/create/:item", CreateTodoItem)
	r.GET("/item/update/:id/:done", UpdateTodoItem)
	r.GET("/item/delete/:id", DeleteTodoItem)

	return r
}

// performRequest is a test helper to make HTTP requests to the test server.
func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestMain sets up the test environment.
// It's run once before all tests in the package.
func TestMain(m *testing.M) {
	// 1. Setup Database Connection
	// This *must* use environment variables for CI (DB_HOST=postgres, etc.)
	// SetupPostgres() // This function (from your main code) should initialize the global 'db'.
	// Forcing a call to a placeholder if not defined, to highlight its necessity.
	// In your actual setup, ensure your real SetupPostgres is called.
	if os.Getenv("CI") != "" { // Simple check if running in a CI-like environment
		fmt.Println("TestMain: Attempting to connect to DB using ENV VARS for CI...")
	}
	// Call your actual DB setup function here. For example:
	// myapp.SetupPostgres() // if SetupPostgres is in myapp package
	// This must initialize the 'db' variable used by emptyTable/displayTable.

	// 2. Setup Router
	router = SetupRoutesForTest() // Use the test-specific router setup.

	// 3. Run Tests
	exitCode := m.Run()

	// 4. Teardown (optional)
	// Example: if db != nil { db.Close() }

	os.Exit(exitCode)
}

// TestItemsGet_EmptyList tests GET /items when the database is empty.
func TestItemsGet_EmptyList(t *testing.T) {
	if db == nil { // Pre-condition check
		t.Fatal("Database connection (db) is nil. Check TestMain and SetupPostgres.")
	}
	emptyTable() // Clear the table before the test.

	w := performRequest(router, "GET", "/items")
	assert.Equal(t, http.StatusOK, w.Code, "Expected HTTP 200 OK")

	// API Handler Requirement: Must return `{"items": []}` (empty JSON array)
	var responseBody map[string][]ListItem
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)

	assert.Nil(t, err, "JSON unmarshalling should succeed")
	if err != nil {
		t.Logf("Response body was: %s", w.Body.String())
	}

	items, exists := responseBody["items"]
	assert.True(t, exists, "Response should contain 'items' key")
	assert.NotNil(t, items, "'items' array should be empty, not nil (e.g. `[]` not `null`)")
	assert.Len(t, items, 0, "There should be 0 items in the list")
}

// TestItemCreate_SingleItem tests POST /item/create/:item for a single item.
func TestItemCreate_SingleItem(t *testing.T) {
	if db == nil {
		t.Fatal("Database connection (db) is nil.")
	}
	emptyTable()

	itemName := "TestItem1"
	// API Handler Requirement: Create handler should take item name, save it, DB assigns ID.
	// Handler should return the created item, including its new ID.
	// The original test expected response key "items" for a single item. Adjust if your API returns differently (e.g. "item").
	w := performRequest(router, "GET", fmt.Sprintf("/item/create/%s", itemName)) // Using GET as per original routes
	assert.Equal(t, http.StatusCreated, w.Code, "Expected HTTP 201 Created")

	var responseBody map[string]ListItem // Assuming response is `{"item": ListItem}` or `{"items": ListItem}`
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.Nil(t, err, "JSON unmarshalling should succeed")
	if err != nil {
		t.Logf("Response body was: %s", w.Body.String())
	}

	// Adjust key if your API returns {"item": ...} instead of {"items": ...} for single create
	createdItem, exists := responseBody["item"] // Changed from "items" to "item" for typical single resource response
	if !exists {                                // Fallback if original "items" key is used
		createdItem, exists = responseBody["items"]
	}
	assert.True(t, exists, "Response should contain the created item under 'item' or 'items' key")

	assert.Equal(t, itemName, createdItem.Item, "Created item's name should match")
	assert.False(t, createdItem.Done, "Newly created item should not be done")
	assert.NotEmpty(t, createdItem.Id, "Created item should have a non-empty ID from the database")
}

// TestItemsCreate_MultipleItems tests creating multiple items and then listing them.
func TestItemsCreate_MultipleItems(t *testing.T) {
	if db == nil {
		t.Fatal("Database connection (db) is nil.")
	}
	emptyTable()

	item1Name := "MultiTest1"
	item2Name := "MultiTest2"

	wCreate1 := performRequest(router, "GET", fmt.Sprintf("/item/create/%s", item1Name))
	assert.Equal(t, http.StatusCreated, wCreate1.Code)
	// Optionally unmarshal and check wCreate1.Body if needed

	wCreate2 := performRequest(router, "GET", fmt.Sprintf("/item/create/%s", item2Name))
	assert.Equal(t, http.StatusCreated, wCreate2.Code)
	// Optionally unmarshal and check wCreate2.Body if needed

	wList := performRequest(router, "GET", "/items")
	assert.Equal(t, http.StatusOK, wList.Code)

	var listResponseBody map[string][]ListItem
	err := json.Unmarshal(wList.Body.Bytes(), &listResponseBody)
	assert.Nil(t, err)

	listedItems, exists := listResponseBody["items"]
	assert.True(t, exists)
	assert.Len(t, listedItems, 2, "Should be 2 items in the list")

	// Check if items exist (order might not be guaranteed unless API sorts)
	foundItem1 := false
	foundItem2 := false
	for _, item := range listedItems {
		if item.Item == item1Name {
			foundItem1 = true
			assert.NotEmpty(t, item.Id)
			assert.False(t, item.Done)
		}
		if item.Item == item2Name {
			foundItem2 = true
			assert.NotEmpty(t, item.Id)
			assert.False(t, item.Done)
		}
	}
	assert.True(t, foundItem1, "Item 1 should be in the list")
	assert.True(t, foundItem2, "Item 2 should be in the list")
}

// TestItemDelete_ExistingItem tests deleting an existing item.
func TestItemDelete_ExistingItem(t *testing.T) {
	if db == nil {
		t.Fatal("Database connection (db) is nil.")
	}
	emptyTable()

	// Create an item to delete and one to keep
	performRequest(router, "GET", "/item/create/ToDelete") // Will get ID "1" (assuming sequence reset)
	performRequest(router, "GET", "/item/create/ToKeep")   // Will get ID "2"

	// Delete item with ID "1"
	// API Handler Requirement: Successful delete should return 200 OK or 204 No Content.
	// If it returns a body, it might be `{"message": "deleted"}`.
	wDelete := performRequest(router, "GET", "/item/delete/1") // ID "1" is assumed
	assert.Contains(t, []int{http.StatusOK, http.StatusNoContent}, wDelete.Code, "Delete should return 200 OK or 204 No Content")

	// Verify it's deleted by trying to list items
	wList := performRequest(router, "GET", "/items")
	assert.Equal(t, http.StatusOK, wList.Code)

	var listResponseBody map[string][]ListItem
	err := json.Unmarshal(wList.Body.Bytes(), &listResponseBody)
	assert.Nil(t, err)

	listedItems, exists := listResponseBody["items"]
	assert.True(t, exists)
	assert.Len(t, listedItems, 1, "Only one item should remain")
	if len(listedItems) == 1 {
		assert.Equal(t, "ToKeep", listedItems[0].Item)
		// assert.Equal(t, "2", listedItems[0].Id) // ID check can be tricky if not predictable
	}
}

// TestItemDelete_NotExistingItem tests deleting a non-existent item.
func TestItemDelete_NotExistingItem(t *testing.T) {
	if db == nil {
		t.Fatal("Database connection (db) is nil.")
	}
	emptyTable()

	// API Handler Requirement: Deleting non-existent item should return 404 Not Found
	// with a body like `{"message": "not found"}`.
	w := performRequest(router, "GET", "/item/delete/999") // ID 999 assumed not to exist
	assert.Equal(t, http.StatusNotFound, w.Code, "Expected HTTP 404 Not Found")

	var responseBody map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.Nil(t, err, "JSON unmarshalling for error message should succeed")
	if err != nil {
		t.Logf("Response body was: %s", w.Body.String())
	}

	message, exists := responseBody["message"]
	assert.True(t, exists, "Error response should contain 'message' key")
	assert.Equal(t, "not found", message, "Error message should be 'not found'")
}

// TestItemUpdate_ExistingItem tests updating an existing item.
func TestItemUpdate_ExistingItem(t *testing.T) {
	if db == nil {
		t.Fatal("Database connection (db) is nil.")
	}
	emptyTable()

	// Create an item
	wCreate := performRequest(router, "GET", "/item/create/ToUpdate")
	assert.Equal(t, http.StatusCreated, wCreate.Code)
	var createRespBody map[string]ListItem
	_ = json.Unmarshal(wCreate.Body.Bytes(), &createRespBody)
	// createdID := createRespBody["item"].Id // Assuming response is `{"item": ...}` and contains ID
	// For simplicity, assuming first created item gets ID "1" due to emptyTable()
	createdID := "1"

	// API Handler Requirement: Update should change the item's 'done' status.
	// Should return 200 OK, possibly with the updated item.
	wUpdate := performRequest(router, "GET", fmt.Sprintf("/item/update/%s/true", createdID))
	assert.Equal(t, http.StatusOK, wUpdate.Code, "Expected HTTP 200 OK for update")

	// Verify by fetching the item or listing all items
	wList := performRequest(router, "GET", "/items")
	assert.Equal(t, http.StatusOK, wList.Code)

	var listResponseBody map[string][]ListItem
	err := json.Unmarshal(wList.Body.Bytes(), &listResponseBody)
	assert.Nil(t, err)

	listedItems, exists := listResponseBody["items"]
	assert.True(t, exists)
	foundUpdated := false
	for _, item := range listedItems {
		if item.Id == createdID {
			assert.Equal(t, "ToUpdate", item.Item)
			assert.True(t, item.Done, "Item should be marked as done after update")
			foundUpdated = true
			break
		}
	}
	assert.True(t, foundUpdated, "Updated item should be found in the list")
}

// TestItemUpdate_NotExistingItem tests updating a non-existent item.
func TestItemUpdate_NotExistingItem(t *testing.T) {
	if db == nil {
		t.Fatal("Database connection (db) is nil.")
	}
	emptyTable()

	// API Handler Requirement: Updating non-existent item should return 404 Not Found
	// with a body like `{"message": "not found"}`.
	w := performRequest(router, "GET", "/item/update/999/true") // ID 999 assumed not to exist
	assert.Equal(t, http.StatusNotFound, w.Code, "Expected HTTP 404 Not Found")

	var responseBody map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.Nil(t, err, "JSON unmarshalling for error message should succeed")
	if err != nil {
		t.Logf("Response body was: %s", w.Body.String())
	}
	message, exists := responseBody["message"]
	assert.True(t, exists, "Error response should contain 'message' key")
	assert.Equal(t, "not found", message, "Error message should be 'not found'")
}
