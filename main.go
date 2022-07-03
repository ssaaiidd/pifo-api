package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	f "github.com/fauna/faunadb-go/v4/faunadb"
)

var (
	port := os.Getenv("PORT")
	secret   = os.Getenv("FAUNA_ENV")
	endpoint = f.Endpoint("https://db.eu.fauna.com")

	dbClient = f.NewFaunaClient(secret, endpoint)

	router = gin.Default()
)

type Note struct {
	Note string `json:"note"`
}

func getNotes(c *gin.Context) {
	res, err := dbClient.Query(
		f.Map(
			f.Paginate(f.Match(f.Index("note"))),
			f.Lambda("X", f.Get(f.Var("X"))),
		))
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "no notes found"})
		panic(err)
	}

	c.IndentedJSON(http.StatusOK, res)
}

func getNotesByID(c *gin.Context) {
	id := c.Param("id")

	res, err := dbClient.Query(
		f.Get(f.RefCollection(f.Collection("notes"), id)),
	)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "note not found"})
		panic(err)
	}

	c.IndentedJSON(http.StatusOK, res)
}

func deleteNote(c *gin.Context) {
	id := c.Param("id")

	res, err := dbClient.Query(
		f.Delete(f.RefCollection(f.Collection("notes"), id)),
	)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "server has a error"})
		panic(err)
	}

	c.IndentedJSON(http.StatusOK, res)
}

func createNote(c *gin.Context) {
	var newNote Note

	if err := c.BindJSON(&newNote); err != nil {
		return
	}

	res, err := dbClient.Query(
		f.Create(f.Collection("notes"), f.Obj{"data": newNote}),
	)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "server has a error"})
		panic(err)
	}

	c.IndentedJSON(http.StatusOK, res)
}

func updateNote(c *gin.Context) {
	id := c.Param("id")

	var newVersionNote Note

	if err := c.BindJSON(&newVersionNote); err != nil {
		return
	}

	res, err := dbClient.Query(
		f.Update(
			f.RefCollection(f.Collection("notes"), id),
			f.Obj{"data": newVersionNote},
		))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "server has a error"})
		panic(err)
	}

	c.IndentedJSON(http.StatusOK, res)
}

func main() {
	router.GET("/notes", getNotes)
	router.GET("/notes/:id", getNotesByID)
	router.GET("/notes/:id/delete", deleteNote)
	router.POST("/notes", createNote)
	router.POST("/notes/:id", updateNote)

	router.Run(":" + port)
}
