package repository

import (
	"context"
	"errors"
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

	expired, err := repo.ExpireOverdue(context.Background(), now)
	if err != nil {
		t.Fatalf("expire overdue: %v", err)
	}
	if expired != 2 {
		t.Errorf("expected 2 expired rows, got %d", expired)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestReviewRepositoryFindActiveByPaperReviewer(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewReviewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "reviews" WHERE paper_id = \$1 AND reviewer_id = \$2 AND status IN \(\$3,\$4\) ORDER BY id DESC,"reviews"\."id" LIMIT \$5`).
		WithArgs(1, 7, constants.ReviewStatusInvited, constants.ReviewStatusAccepted, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "paper_id", "reviewer_id", "status"}).
			AddRow(9, 1, 7, constants.ReviewStatusInvited))

	review, err := repo.FindActiveByPaperReviewer(context.Background(), 1, 7)
	if err != nil {
		t.Fatalf("find active: %v", err)
	}
	if review.ID != 9 || review.Status != constants.ReviewStatusInvited {
		t.Errorf("unexpected active review: %+v", review)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestReviewRepositoryFindActiveByPaperReviewerNotFound(t *testing.T) {
	gormDB, mock := newMockDB(t)
	repo := NewReviewRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "reviews"`).
		WithArgs(1, 7, constants.ReviewStatusInvited, constants.ReviewStatusAccepted, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "paper_id", "reviewer_id", "status"}))

	_, err := repo.FindActiveByPaperReviewer(context.Background(), 1, 7)
	if err == nil || !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
