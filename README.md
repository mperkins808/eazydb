# eazydb

Easy usage of a database in Go.

No code generation used, your IDE can breathe a sigh of relief.

No CLIs required, this is plug and play

**This is very much in alpha**

## Install

```
go get github.com/mperkins808/eazydb/go/pkg/eazydb
```

## Usage

Currently this is only tested with a local postgres instance

Simple docker command to run your own instance

```
docker run --name postgres_container \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=postgres \
  -p 5432:5432 \
  -d postgres
```

### Connecting to a database

```go
c, err := eazydb.NewClient(eazydb.ClientOptions{
    User:     "postgres",
    Password: "postgres",
    Host:     "localhost",
    Port:     "5432",
    Name:     "postgres",
    Type:     eazydb.POSTGRES,
})
```

or if using environment variables

```go
c, err := eazydb.NewClient()
```

### Create a table

Include a json tag in your struct and that field will be created

```go
type Users struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

// Create a table
_, err = c.NewTable("users").Fields(User{}).Key("id", dbtypes.SERIAL).Exec()
if err != nil {
    log.Fatalf("could not create table %v", err)
}

// Shorthand to use later
table := c.Table("users")
```

### Inserting data into a table

Very simple, just parse the struct or []struct and it'll get inserted

```go
// Insert data
users := makeUsers(5000)
metadata, err := table.Add(users).Exec()
if err != nil {
    log.Fatalf("could not insert users table %v", err)
}
```

### Update a field

Again, very easy to do, just defined the fields you want updated

```go
// Update a field
Mat := &User{
    Age: 24,
}
metadata, err = table.Update(Mat).Where(
    *eazydb.String("name").Equals("Mat"),
).Exec()
```

### Deleting a row

Similar to updating fields, except the whole row will be deleted

```go
// Delete
metadata, err = table.Delete().Where(
    *eazydb.Int("age").Equals(40),
).Exec()
```

### Fetching Data

Probably the best part, just parse the structure you want the data as and youre good to go

```go
// Get fields, the best part is parsing directly to an object
var resp []User
_, err = table.Get(User{}).Where(
    *eazydb.Int("age").Equals(24),
).Exec(&resp)
```

### Working example

Theres a working example [here](./go/pkg/eazydb/cmd/main.go). Just make sure you ran the docker command to create a postgress instance.
