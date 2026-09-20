package service

import (
	"context"
	"testing"
	"time"

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

// newReviewTestEnv 构造带一篇论文和一名审稿人的测试环境。
func newReviewTestEnv(t *testing.T) (*fakeStore, *ReviewService, *model.Paper, *model.User) {
	t.Helper()
	store := newFakeStore()
	svc := NewReviewService(store, newTestLogger())
	ctx := context.Background()
	paper := &model.Paper{Title: "外审论文", Status: constants.PaperStatusExternalReview, SubmitterID: 1}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}
	reviewer := &model.User{Username: "rev", Password: "x", RealName: "审稿人", Role: constants.RoleReviewer}
	if err := store.users.Create(ctx, reviewer); err != nil {
		t.Fatalf("create reviewer: %v", err)
	}
	return store, svc, paper, reviewer
}

func TestReviewServiceAssignDuplicateActive(t *testing.T) {
	store, svc, paper, reviewer := newReviewTestEnv(t)
	ctx := context.Background()

	if _, err := svc.Assign(ctx, paper.ID, reviewer.ID); err != nil {
		t.Fatalf("first assign: %v", err)
	}
	if _, err := svc.Assign(ctx, paper.ID, reviewer.ID); err == nil {
		t.Fatalf("expected error: same reviewer must not hold two active tasks")
	}
	reviews, _ := store.reviews.ListByPaper(ctx, paper.ID)
	if len(reviews) != 1 {
		t.Errorf("expected 1 review task, got %d", len(reviews))
	}
}

func TestReviewServiceReinviteAfterDeclined(t *testing.T) {
	store, svc, paper, reviewer := newReviewTestEnv(t)
	ctx := context.Background()

	old := &model.Review{PaperID: paper.ID, ReviewerID: reviewer.ID, Status: constants.ReviewStatusDeclined, Round: 1}
	if err := store.reviews.Create(ctx, old); err != nil {
		t.Fatalf("create declined review: %v", err)
	}
	invite, err := svc.Assign(ctx, paper.ID, reviewer.ID)
	if err != nil {
		t.Fatalf("re-invite after declined: %v", err)
	}
	if invite.ID == old.ID {
		t.Errorf("expected new task record, got old id=%d", invite.ID)
	}
	if invite.Status != constants.ReviewStatusInvited || invite.Round != 1 {
		t.Errorf("unexpected invite: status=%s round=%d", invite.Status, invite.Round)
	}
}

func TestReviewServiceReinviteAfterExpired(t *testing.T) {
	store, svc, paper, reviewer := newReviewTestEnv(t)
	ctx := context.Background()

	past := time.Now().Add(-time.Hour)
	old := &model.Review{PaperID: paper.ID, ReviewerID: reviewer.ID, Status: constants.ReviewStatusInvited, Round: 1, DueDate: &past}
	if err := store.reviews.Create(ctx, old); err != nil {
		t.Fatalf("create overdue review: %v", err)
	}
	invite, err := svc.Assign(ctx, paper.ID, reviewer.ID)
	if err != nil {
		t.Fatalf("re-invite after expired: %v", err)
	}
	oldAfter, _ := store.reviews.FindByID(ctx, old.ID)
	if oldAfter.Status != constants.ReviewStatusExpired {
		t.Errorf("expected old record expired, got %s", oldAfter.Status)
	}
	if invite.Status != constants.ReviewStatusInvited {
		t.Errorf("expected new invite, got %s", invite.Status)
	}
}

func TestReviewServiceRespondExpiredForbidden(t *testing.T) {
	store, svc, paper, reviewer := newReviewTestEnv(t)
	ctx := context.Background()

	past := time.Now().Add(-time.Hour)
	review := &model.Review{PaperID: paper.ID, ReviewerID: reviewer.ID, Status: constants.ReviewStatusInvited, Round: 1, DueDate: &past}
	if err := store.reviews.Create(ctx, review); err != nil {
		t.Fatalf("create review: %v", err)
	}
	if err := svc.Respond(ctx, review.ID, reviewer.ID, true); err == nil {
		t.Fatalf("expected error: expired task must not be responded")
	}
	after, _ := store.reviews.FindByID(ctx, review.ID)
	if after.Status != constants.ReviewStatusExpired {
		t.Errorf("expected expired after blocked respond, got %s", after.Status)
	}
}

func TestReviewServiceSubmitExpiredForbidden(t *testing.T) {
	store, svc, paper, reviewer := newReviewTestEnv(t)
	ctx := context.Background()

	past := time.Now().Add(-time.Hour)
	review := &model.Review{PaperID: paper.ID, ReviewerID: reviewer.ID, Status: constants.ReviewStatusAccepted, Round: 1, DueDate: &past}
	if err := store.reviews.Create(ctx, review); err != nil {
		t.Fatalf("create review: %v", err)
	}
	err := svc.Submit(ctx, review.ID, reviewer.ID, dto.SubmitReviewRequest{
		Decision: constants.ReviewDecisionAccept,
		Comments: "这篇论文整体质量不错，建议录用。",
	})
	if err == nil {
		t.Fatalf("expected error: expired task must not be submitted")
	}
	after, _ := store.reviews.FindByID(ctx, review.ID)
	if after.Status != constants.ReviewStatusExpired {
		t.Errorf("expected expired after blocked submit, got %s", after.Status)
	}
}

func TestReviewServiceSummary(t *testing.T) {
	store, svc, paper, _ := newReviewTestEnv(t)
	ctx := context.Background()

	future := time.Now().Add(24 * time.Hour)
	seed := []model.Review{
		{PaperID: paper.ID, ReviewerID: 1, Status: constants.ReviewStatusCompleted, Round: 1},
		{PaperID: paper.ID, ReviewerID: 2, Status: constants.ReviewStatusDeclined, Round: 1},
		{PaperID: paper.ID, ReviewerID: 3, Status: constants.ReviewStatusExpired, Round: 1},
		{PaperID: paper.ID, ReviewerID: 1, Status: constants.ReviewStatusCompleted, Round: 2},
		{PaperID: paper.ID, ReviewerID: 2, Status: constants.ReviewStatusInvited, Round: 2, DueDate: &future},
		{PaperID: paper.ID, ReviewerID: 4, Status: constants.ReviewStatusAccepted, Round: 2, DueDate: &future},
	}
	for i := range seed {
		if err := store.reviews.Create(ctx, &seed[i]); err != nil {
			t.Fatalf("seed review: %v", err)
		}
	}
	summary, err := svc.Summary(ctx, paper.ID)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Round != 2 {
		t.Errorf("expected round 2, got %d", summary.Round)
	}
	if summary.Completed != 1 || summary.Pending != 1 || summary.InProgress != 1 {
		t.Errorf("unexpected counts: %+v", summary)
	}
	if summary.Declined != 0 || summary.Expired != 0 {
		t.Errorf("history round must not be counted: %+v", summary)
	}
	if summary.CanFinalize {
		t.Errorf("expected can_finalize=false with pending invite")
	}
}
