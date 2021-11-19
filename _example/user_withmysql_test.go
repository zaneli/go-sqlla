// +build withmysql

package example

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mackee/go-sqlla/v2"
	"github.com/ory/dockertest/v3"
)

var db *sql.DB

//go:generate genddl -outpath=./mysql.sql -driver=mysql

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("mysql", "5.7", []string{
		"MYSQL_ROOT_PASSWORD=secret",
		"MYSQL_DATABASE=test",
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		db, err = sql.Open("mysql", fmt.Sprintf("root:secret@(localhost:%s)/test?parseTime=true", resource.GetPort("3306/tcp")))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	schemaFile, err := os.Open("./mysql.sql")
	if err != nil {
		log.Fatal("cannot open schema file error:", err)
	}

	b, err := ioutil.ReadAll(schemaFile)
	if err != nil {
		log.Fatal("cannot read schema file error:", err)
	}

	stmts := strings.Split(string(b), ";")
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		_, err := db.Exec(stmt)
		if err != nil {
			log.Fatal("cannot load schema error:", err)
		}
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestInsertOnDuplicateKeyUpdate__WithMySQL(t *testing.T) {
	ctx := context.Background()

	q1 := NewUserSQL().Insert().
		ValueName("hogehoge").
		ValueRate(3.14).
		ValueIconImage([]byte{}).
		ValueAge(sql.NullInt64{Valid: true, Int64: 17})
	query, args, _ := q1.ToSql()
	t.Logf("query=%s, args=%+v", query, args)
	r1, err := q1.ExecContext(ctx, db)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	time.Sleep(1 * time.Second)

	q2 := NewUserSQL().Insert().
		ValueName("hogehoge").
		ValueAge(sql.NullInt64{Valid: true, Int64: 17}).
		ValueIconImage([]byte{}).
		OnDuplicateKeyUpdate().
		RawValueOnUpdateAge(sqlla.SetMapRawValue("`age` + 1"))
	r2, err := q2.ExecContext(ctx, db)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if r2.Rate != 3.14 {
		t.Fatal("rate is not match:", r2.Rate)
	}
	if r2.Age.Int64 != 18 {
		t.Fatal("age does not incremented:", r2.Age.Int64)
	}
	if r2.UpdatedAt.Time.Unix() <= r1.UpdatedAt.Time.Unix() {
		t.Fatal("updated_at does not updated:", r1.UpdatedAt.Time.Unix(), r2.UpdatedAt.Time.Unix())
	}

}
