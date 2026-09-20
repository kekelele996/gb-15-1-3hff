package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/paperflow/paperflow/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}
	return gormDB, mock
}

func TestUserRepositoryFindByUsername(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewUserRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("author", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "role"}).
			AddRow(1, "author", "hashed", "author"))

	user, err := repo.FindByUsername(context.Background(), "author")
	if err != nil {
		t.Fatalf("find by username: %v", err)
	}
	if user.Username != "author" || user.Role != "author" {
		t.Errorf("unexpected user: %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUserRepositoryFindByUsernameNotFound(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewUserRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("ghost", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.FindByUsername(context.Background(), "ghost")
	if err == nil || !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserRepositoryCreate(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewUserRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	user := &model.User{Username: "alice", Password: "hashed", RealName: "爱丽丝", Role: "author"}
	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("create: %v", err)
	}
	if user.ID != 1 {
		t.Errorf("expected id 1, got %d", user.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
