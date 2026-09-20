package service

import (
	"context"
	"testing"

	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/dto"
	"github.com/paperflow/paperflow/internal/model"
)

func TestReviewServiceRespondAndSubmit(t *testing.T) {
	store := newFakeStore()
	svc := NewReviewService(store, newTestLogger())
	ctx := context.Background()

	paper := &model.Paper{Title: "审稿测试论文", Status: constants.PaperStatusInitialReview, SubmitterID: 1}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}
	reviewer := &model.User{Username: "rev", Password: "x", RealName: "审稿人", Role: constants.RoleReviewer}
	if err := store.users.Create(ctx, reviewer); err != nil {
		t.Fatalf("create reviewer: %v", err)
	}
	review := &model.Review{PaperID: paper.ID, ReviewerID: reviewer.ID, Status: constants.ReviewStatusInvited}
	if err := store.reviews.Create(ctx, review); err != nil {
		t.Fatalf("create review: %v", err)
	}

	if err := svc.Respond(ctx, review.ID, reviewer.ID, true); err != nil {
		t.Fatalf("respond accept: %v", err)
	}
	updated, _ := store.reviews.FindByID(ctx, review.ID)
	if updated.Status != constants.ReviewStatusAccepted {
		t.Errorf("expected accepted, got %s", updated.Status)
	}
	paperAfter, _ := store.papers.FindByID(ctx, paper.ID)
	if paperAfter.Status != constants.PaperStatusExternalReview {
		t.Errorf("expected external_review after accept, got %s", paperAfter.Status)
	}

	if err := svc.Submit(ctx, review.ID, reviewer.ID, dto.SubmitReviewRequest{
		Decision: constants.ReviewDecisionMinorRevision,
		Comments: "实验部分需要补充对比实验，结论建议收敛。",
	}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	done, _ := store.reviews.FindByID(ctx, review.ID)
	if done.Status != constants.ReviewStatusCompleted {
		t.Errorf("expected completed, got %s", done.Status)
	}
	if done.Decision != constants.ReviewDecisionMinorRevision {
		t.Errorf("expected minor_revision, got %s", done.Decision)
	}
	paperFinal, _ := store.papers.FindByID(ctx, paper.ID)
	if paperFinal.Status != constants.PaperStatusRevision {
		t.Errorf("expected revision, got %s", paperFinal.Status)
	}
}

func TestReviewServiceRespondWrongOwner(t *testing.T) {
	store := newFakeStore()
	svc := NewReviewService(store, newTestLogger())
	ctx := context.Background()

	paper := &model.Paper{Title: "P", Status: constants.PaperStatusInitialReview, SubmitterID: 1}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}
	reviewer := &model.User{Username: "rev", Password: "x", RealName: "审稿人", Role: constants.RoleReviewer}
	if err := store.users.Create(ctx, reviewer); err != nil {
		t.Fatalf("create reviewer: %v", err)
	}
	review := &model.Review{PaperID: paper.ID, ReviewerID: reviewer.ID, Status: constants.ReviewStatusInvited}
	if err := store.reviews.Create(ctx, review); err != nil {
		t.Fatalf("create review: %v", err)
	}
	if err := svc.Respond(ctx, review.ID, 999, true); err == nil {
		t.Fatalf("expected permission error for wrong owner")
	}
}
