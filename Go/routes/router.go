package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"full_stack_blog/db"
	"full_stack_blog/models"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func NewRouterWIthDB(database *db.Database) *chi.Mux {
	r := chi.NewRouter()
	// set middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(time.Minute * 10))
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Recoverer)
	// set routes
	r.Get("/posts", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(io.EOF, err) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("Content-Type", "text/charset-utf8")
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len(err.Error())))
			w.Write([]byte(err.Error()))
			return
		}
		filter := bson.D{}
		if year, ok := body["year"].(uint); ok {
			filter = append(filter, bson.E{Key: "year", Value: year})
		}
		if month, ok := body["month"].(uint); ok {
			filter = append(filter, bson.E{Key: "month", Value: month})
		}
		if day, ok := body["day"].(uint); ok {
			filter = append(filter, bson.E{Key: "day", Value: day})
		}
		collection := database.Client.Database("blog").Collection("posts")
		cursor, err := collection.Find(database.Context, filter)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("Content-Type", "text/charset-utf8")
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len(err.Error())))
			w.Write([]byte(err.Error()))
			return
		}
		var posts []models.Post
		for cursor.Next(database.Context) {
			var post models.Post
			if err := cursor.Decode(&post); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Add("Content-Type", "text/charset-utf8")
				w.Header().Add("Content-Length", fmt.Sprintf("%d", len(err.Error())))
				w.Write([]byte(err.Error()))
				return
			}
			posts = append(posts, post)
		}
		if err := cursor.Close(database.Context); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("Content-Type", "text/charset-utf8")
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len(err.Error())))
			w.Write([]byte(err.Error()))
			return
		}
		if err := json.NewEncoder(w).Encode(&posts); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len(err.Error())))
			w.Write([]byte(err.Error()))
		}
	})
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		err := database.Client.Ping(database.Context, readpref.Primary())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("Content-Type", "text/charset-utf8")
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len(err.Error())))
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("Content-Type", "text/charset-utf8")
		w.Header().Add("Content-Length", "4")
		w.Write([]byte("Pong"))
	})
	r.Post("/new", func(w http.ResponseWriter, r *http.Request) {
		var post models.Post
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Header().Add("Content-Type", "text/charset-utf8")
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len(err.Error())))
			w.Write([]byte(err.Error()))
			return
		}
		post.CreatedAt, post.UpdatedAt = time.Now().UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339)
		if post.Day == 0 && post.Year == 0 && post.Month == 0 {
			date := strings.Split(post.CreatedAt, "-")
			year, err := strconv.Atoi(date[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Add("Content-Type", "text/charset-utf8")
				w.Header().Add("Content-Length", fmt.Sprintf("%d", len(err.Error())))
				w.Write([]byte(err.Error()))
				return
			}
			day, err := strconv.Atoi(date[2][:2])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Add("Content-Type", "text/charset-utf8")
				w.Header().Add("Content-Length", fmt.Sprintf("%d", len(err.Error())))
				w.Write([]byte(err.Error()))
				return
			}
			month, err := strconv.Atoi(date[1])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Add("Content-Type", "text/charset-utf8")
				w.Header().Add("Content-Length", fmt.Sprintf("%d", len(err.Error())))
				w.Write([]byte(err.Error()))
				return
			}
			post.Year, post.Month, post.Day = uint(year), uint(month), uint(day)
		}
		collection := database.Client.Database("blog").Collection("posts")
		if _, err := collection.InsertOne(database.Context, post); err != nil {
			if strings.Contains(err.Error(), "duplicate key error") {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Add("Content-Type", "application/json")
				message := "{\"error\":\"A blog post with this date already exists!\"}"
				w.Header().Add("Content-Length", fmt.Sprintf("%d", len(message)))
				w.Write([]byte(message))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("Content-Type", "text/charset-utf8")
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len(err.Error())))
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	r.Post("/edit", func(w http.ResponseWriter, r *http.Request) {
		var post models.Post
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Header().Add("Content-Type", "text/charset-utf8")
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len(err.Error())))
			w.Write([]byte(err.Error()))
			return
		}
		if post.Year == 0 || post.Month == 0 || post.Day == 0 {
			w.WriteHeader(http.StatusBadRequest)
			message := "Request body must contain valid year, month, and day"
			w.Header().Add("Content-Type", "text/charset-utf8")
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len(message)))
			w.Write([]byte(message))
			return
		}
		collection := database.Client.Database("blog").Collection("posts")
		if _, err := collection.UpdateOne(
			database.Context,
			bson.D{
				{Key: "year", Value: post.Year},
				{Key: "month", Value: post.Month},
				{Key: "day", Value: post.Day},
			},
			bson.D{
				{Key: "title", Value: post.Title},
				{Key: "content", Value: post.Content},
			},
		); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("Content-Type", "text/charset-utf8")
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len(err.Error())))
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	r.Delete("/delete", func(w http.ResponseWriter, r *http.Request) {
		var post models.Post
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Header().Add("Content-Type", "text/charset-utf8")
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len(err.Error())))
			w.Write([]byte(err.Error()))
			return
		}
		if post.Year == 0 || post.Month == 0 || post.Day == 0 {
			w.WriteHeader(http.StatusBadRequest)
			message := "Request body must contain valid year, month, and day"
			w.Header().Add("Content-Type", "text/charset-utf8")
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len(message)))
			w.Write([]byte(message))
			return
		}
		collection := database.Client.Database("blog").Collection("posts")
		if _, err := collection.DeleteOne(
			database.Context,
			bson.D{
				{Key: "year", Value: post.Year},
				{Key: "month", Value: post.Month},
				{Key: "day", Value: post.Day},
			}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("Content-Type", "text/charset-utf8")
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len(err.Error())))
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	return r
}
