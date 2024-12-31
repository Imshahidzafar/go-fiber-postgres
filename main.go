package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"github.com/gofiber/fiber/v2"
	"github.com/imshahidzafar/go-fiber-postgres/models"
	"github.com/imshahidzafar/go-fiber-postgres/storage"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Repoository struct {
	DB *gorm.DB
}

func (r *Repoository) CreateBook(context *fiber.Ctx) error {
	book := models.Books{}

	err := context.BodyParser(&book)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	err = r.DB.Create(&book).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not create book"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book has been added",
	})
	return nil
}

func (r *Repoository) GetBooks(context *fiber.Ctx) error {
	bookModels := &[]models.Books{}

	err := r.DB.Find(bookModels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get books"})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "books fetched successfully", "data": bookModels})
	return nil
}

func (r *Repoository) DeleteBook(context *fiber.Ctx) error {
	bookModel := models.Books{}
	id := context.Params("id")
	bookID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "invalid book ID",
		})
		return err
	}

	err = r.DB.Delete(&bookModel, uint(bookID)).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not delete book",
		})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book has been deleted",
	})
	return nil
}

func (r *Repoository) GetBookByID(context *fiber.Ctx) error {
	id := context.Params("id")
	bookModel := &models.Books{}
	bookID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "invalid book ID"})
		return err
	}

	err = r.DB.Where("id = ?", uint(bookID)).First(bookModel).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "could not get the book"})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "book id fetched successfully", "data": bookModel})
	return nil
}

func (r *Repoository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_books", r.CreateBook)
	api.Delete("/delete_book/:id", r.DeleteBook)
	api.Get("/books/:id", r.GetBookByID)
	api.Get("/books", r.GetBooks)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}
	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	} else {
		log.Println("Database connection successful")
	}

	err = models.MigrateBooks(db)
	if err != nil {
		log.Fatal("Could not migrate db")
	}

	r := Repoository{
		DB: db,
	}

	app := fiber.New()
	r.SetupRoutes(app)
	log.Println("Server is running on http://localhost:3000")
	app.Listen(":3000")
}
