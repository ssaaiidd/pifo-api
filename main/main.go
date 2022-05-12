package main

import (
	f "github.com/fauna/faunadb-go/faunadb"
	"log"
	"os"
)

var (
	secret =  os.Getenv("FAUNA_ENV")
	endpoint = f.Endpoint("https://db.fauna.com")

	adminClient = f.NewFaunaClient(secret, endpoint)

	dbName = "pifo"
)

/*
 *  Check for the existence of the database. If it exists return true.
 *  If database does not exist create it.
 */
func createDatabase() {
	res, err := adminClient.Query(
		f.If(
			f.Exists(f.Database(dbName)),
			true,
			f.CreateDatabase(f.Obj{"name": dbName})))

	if err != nil {
		panic(err)
	}

	if res != f.BooleanV(true) {
		log.Printf("Created Database: %s\n %s", dbName, res)
	} else {
		log.Printf("Database: %s, Already Exists\n %s", dbName, res)
	}
}

/*
 * Create a Database specific client that we can use to create objects within
 * the target database. In this case we will give the client the role of "server"
 * which will allow us create/delete/read/write access to all objects in the database.
 */
func getDbClient() (dbClient *f.FaunaClient) {
	var res f.Value
	var err error

	var secret string

	res, err = adminClient.Query(
		f.CreateKey(f.Obj{
			"database": f.Database(dbName),
			"role":     "server"}))

	if err != nil {
		panic(err)
	}

	err = res.At(f.ObjKey("secret")).Get(&secret)

	if err != nil {
		panic(err)
	}

	log.Printf("Database: %s, specifc key: %s\n%s", dbName, secret, res)

	dbClient = adminClient.NewSessionClient(secret)

	return
}

/*
 *  Check for the existence of the class. If it exists return true.
 *  If class does not exist create it.
 */
func createClass( dbClient *f.FaunaClient, className string) {

	res, err := dbClient.Query(
		f.If(
			f.Exists(f.Class(className)),
			true,
			f.CreateClass(f.Obj{"name": className})))

	if err != nil {
		panic(err)
	}

	if res != f.BooleanV(true) {
		log.Printf("Created Class: %s\n %s", className, res)
	} else {
		log.Printf("Class: %s, Already Exists\n %s", className, res)
	}
}

/*
 * Create a new instance of a class with "id" and "name"
 */
func createInstance(dbClient *f.FaunaClient, className string, id int, name string) {
	var res f.Value
	var err error

	var ref f.RefV

	res, err = dbClient.Query(
		f.Create(f.Class(className), f.Obj{"data": f.Obj{"id": id, "name": name}}))

	if err != nil {
		log.Printf("Instance Existed '%s': %v : %s", className, id, name)
		return
	}

	if err = res.At(f.ObjKey("ref")).Get(&ref); err == nil {
		log.Printf("Instance Created '%s': %v : %s", className, id, ref)
	} else {
		panic(err)
	}

	/*
	 * Retrieve the record using the reference.
	 */
	res, err = dbClient.Query(f.Select(f.Arr{"data", "name"}, f.Get(ref)))

	if err != nil {
		panic(err)
	}
	log.Printf("Read by @ref %s: %v : %s", className, id, res)

}

/*
 * Create an index to access customer records by id
 */
func createIndex(dbClient *f.FaunaClient, indexName string, className string, primaryKey string) {

	res, err := dbClient.Query(
		f.If(
			f.Exists(f.Index(indexName)),
			true,
			f.CreateIndex(f.Obj{
				"name":   indexName,
				"source": f.Class(className),
				"unique": true,
				"terms":  f.Obj{"field": f.Arr{"data", primaryKey}}})))

	if err != nil {
		panic(err)
	}

	if res != f.BooleanV(true) {
		log.Printf("Created Index: %s\n %s", indexName, res)
	} else {
		log.Printf("Index: %s, Already Exists\n %s", indexName, res)
	}

}

/*
 * Retrieve the record using an index and the index value.
 */
func getInstanceByPrimaryKey(dbClient *f.FaunaClient, indexName string, primaryKey int) {

	res, err := dbClient.Query(
		f.Select(f.Arr{"data", "name"},
			f.Get(f.MatchTerm(f.Index(indexName), primaryKey))))

	if err != nil {
		panic(err)
	}
	log.Printf("Read by Primary Key %s: %v : %s", indexName, primaryKey, res)

}

func main() {
	createDatabase()

	dbClient := getDbClient()

	className := "Notes"
	createClass(dbClient, className)

	indexName := "note_key"
	primaryKey := "id"
	createIndex(dbClient, indexName, className, primaryKey)

	createInstance(dbClient, className, 1, "Notes 1")
	createInstance(dbClient, className, 2, "Notes 2")
	createInstance(dbClient, className, 3, "Notes 3")
	createInstance(dbClient, className, 4, "Notes 4")

	getInstanceByPrimaryKey(dbClient, indexName, 1)
	getInstanceByPrimaryKey(dbClient, indexName, 2)
	getInstanceByPrimaryKey(dbClient, indexName, 3)
	getInstanceByPrimaryKey(dbClient, indexName, 4)
}

1
