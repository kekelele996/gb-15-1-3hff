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

// 邀请超过截止日期后旧记录失效，审稿人不得再回应。
func TestReviewServiceRespondExpiredInvite(t *testing.T) {
	store := newFakeStore()
	svc := NewReviewService(store, newTestLogger())
	ctx := context.Background()

	paper := &model.Paper{Title: "超期论文", Status: constants.PaperStatusExternalReview, SubmitterID: 1}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}
	reviewer := &model.User{Username: "rev", Password: "x", RealName: "审稿人", Role: constants.RoleReviewer}
	if err := store.users.Create(ctx, reviewer); err != nil {
		t.Fatalf("create reviewer: %v", err)
	}
	past := time.Now().Add(-24 * time.Hour)
	review := &model.Review{PaperID: paper.ID, ReviewerID: reviewer.ID, Status: constants.ReviewStatusInvited, DueDate: &past}
	if err := store.reviews.Create(ctx, review); err != nil {
		t.Fatalf("create review: %v", err)
	}
	if err := svc.Respond(ctx, review.ID, reviewer.ID, true); err == nil {
		t.Fatalf("expected error when responding to expired invite")
	}
	updated, _ := store.reviews.FindByID(ctx, review.ID)
	if updated.Status != constants.ReviewStatusExpired {
		t.Errorf("expected expired after sweep, got %s", updated.Status)
	}
}

// 已超期的审稿中任务不得再提交评审意见。
func TestReviewServiceSubmitExpired(t *testing.T) {
	store := newFakeStore()
	svc := NewReviewService(store, newTestLogger())
	ctx := context.Background()

	paper := &model.Paper{Title: "超期提交论文", Status: constants.PaperStatusExternalReview, SubmitterID: 1}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}
	reviewer := &model.User{Username: "rev", Password: "x", RealName: "审稿人", Role: constants.RoleReviewer}
	if err := store.users.Create(ctx, reviewer); err != nil {
		t.Fatalf("create reviewer: %v", err)
	}
	past := time.Now().Add(-24 * time.Hour)
	review := &model.Review{PaperID: paper.ID, ReviewerID: reviewer.ID, Status: constants.ReviewStatusAccepted, DueDate: &past}
	if err := store.reviews.Create(ctx, review); err != nil {
		t.Fatalf("create review: %v", err)
	}
	err := svc.Submit(ctx, review.ID, reviewer.ID, dto.SubmitReviewRequest{
		Decision: constants.ReviewDecisionAccept,
		Comments: "论文质量良好，建议直接录用。",
	})
	if err == nil {
		t.Fatalf("expected error when submitting expired review")
	}
	updated, _ := store.reviews.FindByID(ctx, review.ID)
	if updated.Status != constants.ReviewStatusExpired {
		t.Errorf("expected expired after sweep, got %s", updated.Status)
	}
}

// 同一审稿人不能同时存在两条有效任务；旧邀请被拒绝后编辑可补邀，旧任务不得再回应。
func TestReviewServiceAssignDuplicateAndReinvite(t *testing.T) {
	store := newFakeStore()
	svc := NewReviewService(store, newTestLogger())
	ctx := context.Background()

	paper := &model.Paper{Title: "补邀论文", Status: constants.PaperStatusExternalReview, SubmitterID: 1, Version: 1}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}
	reviewer := &model.User{Username: "rev", Password: "x", RealName: "审稿人", Role: constants.RoleReviewer}
	if err := store.users.Create(ctx, reviewer); err != nil {
		t.Fatalf("create reviewer: %v", err)
	}

	first, err := svc.Assign(ctx, paper.ID, reviewer.ID)
	if err != nil {
		t.Fatalf("first assign: %v", err)
	}
	if _, err := svc.Assign(ctx, paper.ID, reviewer.ID); err == nil {
		t.Fatalf("expected duplicate active invite to be rejected")
	}
	if err := svc.Respond(ctx, first.ID, reviewer.ID, false); err != nil {
		t.Fatalf("decline invite: %v", err)
	}
	second, err := svc.Assign(ctx, paper.ID, reviewer.ID)
	if err != nil {
		t.Fatalf("re-invite after decline: %v", err)
	}
	if second.ID == first.ID {
		t.Errorf("expected new invite record after decline")
	}
	if err := svc.Respond(ctx, first.ID, reviewer.ID, true); err == nil {
		t.Fatalf("expected old declined invite to be unanswerable")
	}
	if err := svc.Respond(ctx, second.ID, reviewer.ID, true); err != nil {
		t.Fatalf("respond to new invite: %v", err)
	}
}

// 外审进度汇总：各状态人数统计与终审门槛判定。
func TestReviewServiceSummarize(t *testing.T) {
	store := newFakeStore()
	svc := NewReviewService(store, newTestLogger())
	ctx := context.Background()

	paper := &model.Paper{Title: "汇总论文", Status: constants.PaperStatusExternalReview, SubmitterID: 1, Version: 2}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}
	seed := []model.Review{
		{PaperID: paper.ID, ReviewerID: 1, Status: constants.ReviewStatusCompleted, Round: 1},
		{PaperID: paper.ID, ReviewerID: 1, Status: constants.ReviewStatusCompleted, Round: 2},
		{PaperID: paper.ID, ReviewerID: 2, Status: constants.ReviewStatusCompleted, Round: 1},
		{PaperID: paper.ID, ReviewerID: 3, Status: constants.ReviewStatusInvited, Round: 2},
		{PaperID: paper.ID, ReviewerID: 4, Status: constants.ReviewStatusDeclined, Round: 1},
		{PaperID: paper.ID, ReviewerID: 5, Status: constants.ReviewStatusExpired, Round: 1},
	}
	for i := range seed {
		if err := store.reviews.Create(ctx, &seed[i]); err != nil {
			t.Fatalf("create review: %v", err)
		}
	}
	summary, err := svc.Summarize(ctx, paper.ID)
	if err != nil {
		t.Fatalf("summarize: %v", err)
	}
	if summary.Completed != 3 || summary.CompletedReviewers != 2 {
		t.Errorf("expected completed=3 reviewers=2, got %d/%d", summary.Completed, summary.CompletedReviewers)
	}
	if summary.Invited != 1 || summary.Declined != 1 || summary.Expired != 1 {
		t.Errorf("unexpected counts: %+v", summary)
	}
	if summary.CanFinalize {
		t.Errorf("expected can_finalize=false while invited pending")
	}
	if len(summary.BlockReasons) == 0 {
		t.Errorf("expected block reasons")
	}
}
