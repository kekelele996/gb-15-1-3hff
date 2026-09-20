package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/paperflow/paperflow/internal/model"
	"gorm.io/gorm"
)

func TestPaperRepositoryListByStatus(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewPaperRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "papers" WHERE status = $1`)).
		WithArgs("accepted").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT .* FROM "papers"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "status", "submitter_id"}).
			AddRow(1, "Accepted Paper", "accepted", 7))
	mock.ExpectQuery(`SELECT .* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(7, "author"))

	items, total, err := repo.List(context.Background(), model.PaperFilter{Status: "accepted"}, 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected total=1 len=1, got %d/%d", total, len(items))
	}
	if items[0].Submitter.ID != 7 {
		t.Errorf("expected preloaded submitter id 7, got %d", items[0].Submitter.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPaperRepositoryFindByIDNotFound(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewPaperRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "papers" WHERE "papers"."id" = $1 ORDER BY "papers"."id" LIMIT $2`)).
		WithArgs(99, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := repo.FindByID(context.Background(), 99)
	if err == nil || !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
