package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Product represents a product in the inventory system
type Product struct {
	ID       int64
	Name     string
	Price    float64
	Quantity int
	Category string
}

// ProductStore manages product operations
type ProductStore struct {
	db *sql.DB
}

// NewProductStore creates a new ProductStore with the given database connection
func NewProductStore(db *sql.DB) *ProductStore {
	return &ProductStore{db: db}
}

// InitDB sets up a new SQLite database and creates the products table
func InitDB(dbPath string) (*sql.DB, error) {
	// TODO: Open a SQLite database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// TODO: Create the products table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS products(id INTEGER PRIMARY KEY, name TEXT, price REAL, quantity INTEGER, category TEXT)")
	if err != nil {
		return nil, err
	}

	// The table should have columns: id, name, price, quantity, category
	return db, nil
}

// CreateProduct adds a new product to the database
func (ps *ProductStore) CreateProduct(product *Product) (*Product, error) {
	// TODO: Insert the product into the database
	results, err := ps.db.Exec("INSERT INTO products (name, price, quantity, category) VALUES (?, ?, ?, ?)", product.Name, product.Price, product.Quantity, product.Category)
	if err != nil {
		return &Product{}, err
	}

	// TODO: Update the product.ID with the database-generated ID
	id, err := results.LastInsertId()
	if err != nil {
		return &Product{}, err
	}
	product.ID = id

	return product, nil
}

// GetProduct retrieves a product by ID
func (ps *ProductStore) GetProduct(id int64) (*Product, error) {
	// TODO: Query the database for a product with the given ID
	results, err := ps.db.Query("SELECT * FROM products WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	// TODO: Return a Product struct populated with the data or an error if not found
	p := &Product{}
	err = results.Scan(&p.ID, &p.Name, &p.Price, &p.Quantity, &p.Category)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product with ID %d not found", id)
		}
		return nil, err
	}

	return p, nil
}

// UpdateProduct updates an existing product
func (ps *ProductStore) UpdateProduct(product *Product) error {
	// TODO: Update the product in the database
	// TODO: Return an error if the product doesn't exist
	return errors.New("not implemented")
}

// DeleteProduct removes a product by ID
func (ps *ProductStore) DeleteProduct(id int64) error {
	// TODO: Delete the product from the database
	// TODO: Return an error if the product doesn't exist
	return errors.New("not implemented")
}

// ListProducts returns all products with optional filtering by category
func (ps *ProductStore) ListProducts(category string) ([]*Product, error) {
	// TODO: Query the database for products
	// TODO: If category is not empty, filter by category
	// TODO: Return a slice of Product pointers
	return nil, errors.New("not implemented")
}

// BatchUpdateInventory updates the quantity of multiple products in a single transaction
func (ps *ProductStore) BatchUpdateInventory(updates map[int64]int) error {
	// TODO: Start a transaction
	// TODO: For each product ID in the updates map, update its quantity
	// TODO: If any update fails, roll back the transaction
	// TODO: Otherwise, commit the transaction
	return errors.New("not implemented")
}

func main() {
	db, err := InitDB("./sqlite.db")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	product := &Product{
		ID:       1231,
		Name:     "First item",
		Price:    20.2,
		Quantity: 2,
		Category: "sales",
	}

	store := NewProductStore(db)

	product, err = store.CreateProduct(product)
	if err != nil {
		panic(err)
	}

	fmt.Print(product)

	getProduct, err := store.GetProduct(1)
	if err != nil {
		panic(err)
	}

	fmt.Println(getProduct)
}
