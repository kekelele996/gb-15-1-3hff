package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/paperflow/paperflow/internal/constants"
)

func TestReviewRepositoryExpireOverdue(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewReviewRepository(gormDB)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "reviews" SET "status"=\$1,"updated_at"=\$2 WHERE status IN \(\$3,\$4\) AND due_date IS NOT NULL AND due_date < \$5`).
		WithArgs(constants.ReviewStatusExpired, now, constants.ReviewStatusInvited, constants.ReviewStatusAccepted, now).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	affected, err := repo.ExpireOverdue(context.Background(), now)
	if err != nil {
		t.Fatalf("expire overdue: %v", err)
	}
	if affected != 2 {
		t.Errorf("expected 2 expired, got %d", affected)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestReviewRepositoryExpireOpenByPaper(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewReviewRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "reviews" SET "status"=\$1,"updated_at"=\$2 WHERE paper_id = \$3 AND status IN \(\$4,\$5\)`).
		WithArgs(constants.ReviewStatusExpired, sqlmock.AnyArg(), uint(7), constants.ReviewStatusInvited, constants.ReviewStatusAccepted).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	affected, err := repo.ExpireOpenByPaper(context.Background(), 7)
	if err != nil {
		t.Fatalf("expire open by paper: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 expired, got %d", affected)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestReviewRepositoryMaxRoundByPaper(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewReviewRepository(gormDB)

	mock.ExpectQuery(`SELECT COALESCE\(MAX\(round\), 0\) FROM "reviews" WHERE paper_id = \$1`).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(3))

	round, err := repo.MaxRoundByPaper(context.Background(), 7)
	if err != nil {
		t.Fatalf("max round: %v", err)
	}
	if round != 3 {
		t.Errorf("expected round 3, got %d", round)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestReviewRepositoryFindActiveByPaperReviewer(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewReviewRepository(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reviews" WHERE paper_id = $1 AND reviewer_id = $2 AND round = $3 AND status IN ($4,$5,$6) ORDER BY "reviews"."id" LIMIT $7`)).
		WithArgs(7, 9, 2, constants.ReviewStatusInvited, constants.ReviewStatusAccepted, constants.ReviewStatusCompleted, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "paper_id", "reviewer_id", "status", "round"}).
			AddRow(5, 7, 9, constants.ReviewStatusInvited, 2))

	review, err := repo.FindActiveByPaperReviewer(context.Background(), 7, 9, 2)
	if err != nil {
		t.Fatalf("find active: %v", err)
	}
	if review.ID != 5 || review.Round != 2 {
		t.Errorf("unexpected review: %+v", review)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestReviewRepositoryFindActiveByPaperReviewerNotFound(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewReviewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "reviews" WHERE paper_id = \$1 AND reviewer_id = \$2 AND round = \$3`).
		WithArgs(7, 9, 2, constants.ReviewStatusInvited, constants.ReviewStatusAccepted, constants.ReviewStatusCompleted, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindActiveByPaperReviewer(context.Background(), 7, 9, 2)
	if err == nil || !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
