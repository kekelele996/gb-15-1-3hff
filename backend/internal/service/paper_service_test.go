package service

import (
	"context"
	"testing"
	"time"

	"github.com/paperflow/paperflow/internal/constants"
	"github.com/paperflow/paperflow/internal/dto"
	"github.com/paperflow/paperflow/internal/model"
)

func newPaperTestEnv(t *testing.T) (*fakeStore, *PaperService) {
	t.Helper()
	store := newFakeStore()
	cfg := newTestConfig()
	plagiarismSvc := NewPlagiarismService(store, cfg, newTestLogger())
	svc := NewPaperService(store, plagiarismSvc, cfg, newTestLogger())
	return store, svc
}

func TestPaperServiceCreate(t *testing.T) {
	store, svc := newPaperTestEnv(t)
	ctx := context.Background()

	paper, err := svc.Create(ctx, 1, dto.CreatePaperRequest{
		Title:       "测试论文",
		Abstract:    "这是一篇用于单元测试的论文摘要内容，长度满足校验要求。",
		Keywords:    "测试,论文",
		Subject:     "computer",
		AuthorsMeta: `[{"name":"张三","institution":"某大学"}]`,
		FileKey:     "papers/uuid.pdf",
		FileName:    "paper.pdf",
	})
	if err != nil {
		t.Fatalf("create paper: %v", err)
	}
	if paper.Status != constants.PaperStatusSubmitted {
		t.Errorf("expected submitted, got %s", paper.Status)
	}
	if paper.Similarity <= 0 || paper.Similarity > 30 {
		t.Errorf("expected similarity in (0,30], got %f", paper.Similarity)
	}
	check, err := store.plagiarism.FindByPaper(ctx, paper.ID)
	if err != nil {
		t.Fatalf("plagiarism check not found: %v", err)
	}
	if check.Status != constants.PlagiarismStatusCompleted {
		t.Errorf("expected completed, got %s", check.Status)
	}
	if check.Report == "" {
		t.Errorf("expected report not empty")
	}
}

func TestPaperServiceInitialReview(t *testing.T) {
	store, svc := newPaperTestEnv(t)
	ctx := context.Background()

	paper, err := svc.Create(ctx, 1, dto.CreatePaperRequest{
		Title:       "待初审论文",
		Abstract:    "这是一篇用于单元测试的论文摘要内容，长度满足校验要求。",
		Keywords:    "测试,初审",
		Subject:     "computer",
		AuthorsMeta: `[{"name":"张三","institution":"某大学"}]`,
		FileKey:     "papers/uuid.pdf",
		FileName:    "paper.pdf",
	})
	if err != nil {
		t.Fatalf("create paper: %v", err)
	}
	reviewer := &model.User{Username: "reviewer1", Password: "x", RealName: "审稿人1", Role: constants.RoleReviewer}
	if err := store.users.Create(ctx, reviewer); err != nil {
		t.Fatalf("create reviewer: %v", err)
	}

	t.Run("pass assigns reviewer", func(t *testing.T) {
		updated, err := svc.InitialReview(ctx, 2, paper.ID, dto.InitialReviewRequest{
			Pass: true, Reason: "格式规范", ReviewerID: reviewer.ID,
		})
		if err != nil {
			t.Fatalf("initial review pass: %v", err)
		}
		if updated.Status != constants.PaperStatusInitialReview {
			t.Errorf("expected initial_review, got %s", updated.Status)
		}
		invites, err := store.reviews.ListByPaper(ctx, paper.ID)
		if err != nil || len(invites) != 1 {
			t.Fatalf("expected 1 invite, got %d (err=%v)", len(invites), err)
		}
		if invites[0].ReviewerID != reviewer.ID || invites[0].Status != constants.ReviewStatusInvited {
			t.Errorf("unexpected invite: %+v", invites[0])
		}
	})

	t.Run("reject returns to author", func(t *testing.T) {
		paper2, err := svc.Create(ctx, 1, dto.CreatePaperRequest{
			Title:       "待拒论文",
			Abstract:    "这是一篇用于单元测试的论文摘要内容，长度满足校验要求。",
			Keywords:    "测试,拒稿",
			Subject:     "education",
			AuthorsMeta: `[{"name":"张三","institution":"某大学"}]`,
			FileKey:     "papers/uuid2.pdf",
			FileName:    "paper2.pdf",
		})
		if err != nil {
			t.Fatalf("create paper2: %v", err)
		}
		updated, err := svc.InitialReview(ctx, 2, paper2.ID, dto.InitialReviewRequest{
			Pass: false, Reason: "选题不符",
		})
		if err != nil {
			t.Fatalf("initial review reject: %v", err)
		}
		if updated.Status != constants.PaperStatusRejected {
			t.Errorf("expected rejected, got %s", updated.Status)
		}
		if updated.InitialReviewComment != "选题不符" {
			t.Errorf("expected reason, got %s", updated.InitialReviewComment)
		}
	})

	t.Run("wrong status rejected", func(t *testing.T) {
		paper3, err := svc.Create(ctx, 1, dto.CreatePaperRequest{
			Title:       "状态错误论文",
			Abstract:    "这是一篇用于单元测试的论文摘要内容，长度满足校验要求。",
			Keywords:    "测试,状态",
			Subject:     "physics",
			AuthorsMeta: `[{"name":"张三","institution":"某大学"}]`,
			FileKey:     "papers/uuid3.pdf",
			FileName:    "paper3.pdf",
		})
		if err != nil {
			t.Fatalf("create paper3: %v", err)
		}
		paper3.Status = constants.PaperStatusExternalReview
		if err := store.papers.Update(ctx, paper3); err != nil {
			t.Fatalf("set external_review: %v", err)
		}
		if _, err := svc.InitialReview(ctx, 2, paper3.ID, dto.InitialReviewRequest{Pass: true, ReviewerID: reviewer.ID}); err == nil {
			t.Fatalf("expected error for invalid state transition")
		}
	})
}

func TestPaperServiceRevise(t *testing.T) {
	store, svc := newPaperTestEnv(t)
	ctx := context.Background()

	paper, err := svc.Create(ctx, 1, dto.CreatePaperRequest{
		Title:       "修稿论文",
		Abstract:    "这是一篇用于单元测试的论文摘要内容，长度满足校验要求。",
		Keywords:    "测试,修稿",
		Subject:     "computer",
		AuthorsMeta: `[{"name":"张三","institution":"某大学"}]`,
		FileKey:     "papers/uuid.pdf",
		FileName:    "paper.pdf",
	})
	if err != nil {
		t.Fatalf("create paper: %v", err)
	}
	paper.Status = constants.PaperStatusRevision
	if err := store.papers.Update(ctx, paper); err != nil {
		t.Fatalf("set revision: %v", err)
	}

	updated, err := svc.Revise(ctx, 1, paper.ID, dto.ReviseRequest{
		FileKey:        "papers/uuid-v2.pdf",
		FileName:       "paper-v2.pdf",
		ResponseLetter: "已逐条回复审稿意见：1）已补充实验对比；2）已修正格式问题。",
	})
	if err != nil {
		t.Fatalf("revise: %v", err)
	}
	if updated.Version != 2 {
		t.Errorf("expected version 2, got %d", updated.Version)
	}
	if updated.Status != constants.PaperStatusExternalReview {
		t.Errorf("expected external_review, got %s", updated.Status)
	}
	revisions, err := store.revisions.ListByPaper(ctx, paper.ID)
	if err != nil || len(revisions) != 1 {
		t.Fatalf("expected 1 revision, got %d (err=%v)", len(revisions), err)
	}
	if revisions[0].Version != 2 {
		t.Errorf("expected revision version 2, got %d", revisions[0].Version)
	}
}

// 修稿后按仍在有效期内的审稿人开启新一轮匿名评审：旧有效任务失效、旧意见只读保留。
func TestPaperServiceReviseOpensNewReviewRound(t *testing.T) {
	store, svc := newPaperTestEnv(t)
	ctx := context.Background()

	paper := &model.Paper{
		Title: "新一轮评审论文", Status: constants.PaperStatusRevision,
		SubmitterID: 1, Version: 1,
	}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}
	future := time.Now().Add(7 * 24 * time.Hour)
	seed := []model.Review{
		{PaperID: paper.ID, ReviewerID: 11, Status: constants.ReviewStatusCompleted, Round: 1, Decision: constants.ReviewDecisionMajorRevision, Comments: "旧意见只读保留"},
		{PaperID: paper.ID, ReviewerID: 12, Status: constants.ReviewStatusAccepted, Round: 1, DueDate: &future},
		{PaperID: paper.ID, ReviewerID: 13, Status: constants.ReviewStatusDeclined, Round: 1},
	}
	for i := range seed {
		if err := store.reviews.Create(ctx, &seed[i]); err != nil {
			t.Fatalf("create review: %v", err)
		}
	}

	if _, err := svc.Revise(ctx, 1, paper.ID, dto.ReviseRequest{
		FileKey: "papers/v2.pdf", FileName: "v2.pdf", ResponseLetter: "已逐条回复审稿意见并修改。",
	}); err != nil {
		t.Fatalf("revise: %v", err)
	}

	reviews, err := store.reviews.ListByPaper(ctx, paper.ID)
	if err != nil {
		t.Fatalf("list reviews: %v", err)
	}
	byReviewer := map[uint][]model.Review{}
	for _, r := range reviews {
		byReviewer[r.ReviewerID] = append(byReviewer[r.ReviewerID], r)
	}
	// 已完成审稿人：旧意见保留为 completed，新增第 2 轮邀请。
	if got := len(byReviewer[11]); got != 2 {
		t.Fatalf("reviewer 11: expected 2 records, got %d", got)
	}
	old11 := findByRound(byReviewer[11], 1)
	if old11 == nil || old11.Status != constants.ReviewStatusCompleted || old11.Comments != "旧意见只读保留" {
		t.Errorf("reviewer 11: old opinion should stay completed read-only, got %+v", old11)
	}
	// 审稿中审稿人：旧任务失效为 expired，新增第 2 轮邀请。
	if got := len(byReviewer[12]); got != 2 {
		t.Fatalf("reviewer 12: expected 2 records, got %d", got)
	}
	old12 := findByRound(byReviewer[12], 1)
	if old12 == nil || old12.Status != constants.ReviewStatusExpired {
		t.Errorf("reviewer 12: old active task should be expired, got %+v", old12)
	}
	// 已拒绝审稿人：不自动续邀。
	if got := len(byReviewer[13]); got != 1 {
		t.Fatalf("reviewer 13: expected no re-invite, got %d records", got)
	}
	// 新邀请均为第 2 轮待接受。
	for _, rid := range []uint{11, 12} {
		invite := findByRound(byReviewer[rid], 2)
		if invite == nil || invite.Status != constants.ReviewStatusInvited || invite.DueDate == nil {
			t.Errorf("reviewer %d: expected round-2 invited task with due date, got %+v", rid, invite)
		}
	}
}

func findByRound(reviews []model.Review, round int) *model.Review {
	for i := range reviews {
		if reviews[i].Round == round {
			return &reviews[i]
		}
	}
	return nil
}

// 终审门槛：有效完成评审不足两人或存在待处理邀请时阻止终审。
func TestPaperServiceFinalDecisionGate(t *testing.T) {
	store, svc := newPaperTestEnv(t)
	ctx := context.Background()

	paper := &model.Paper{
		Title: "终审门槛论文", Status: constants.PaperStatusExternalReview,
		SubmitterID: 1, Version: 1,
	}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}
	decision := dto.FinalDecisionRequest{Decision: constants.PaperStatusAccepted, Comment: "同意录用"}

	// 仅 1 人完成：阻止。
	r1 := &model.Review{PaperID: paper.ID, ReviewerID: 21, Status: constants.ReviewStatusCompleted, Round: 1}
	if err := store.reviews.Create(ctx, r1); err != nil {
		t.Fatalf("create review: %v", err)
	}
	if _, err := svc.FinalDecision(ctx, 2, paper.ID, decision); err == nil {
		t.Fatalf("expected final decision blocked with only 1 completed review")
	}

	// 2 人完成但存在待接受邀请：阻止。
	r2 := &model.Review{PaperID: paper.ID, ReviewerID: 22, Status: constants.ReviewStatusCompleted, Round: 1}
	r3 := &model.Review{PaperID: paper.ID, ReviewerID: 23, Status: constants.ReviewStatusInvited, Round: 1}
	if err := store.reviews.Create(ctx, r2); err != nil {
		t.Fatalf("create review: %v", err)
	}
	if err := store.reviews.Create(ctx, r3); err != nil {
		t.Fatalf("create review: %v", err)
	}
	if _, err := svc.FinalDecision(ctx, 2, paper.ID, decision); err == nil {
		t.Fatalf("expected final decision blocked with pending invite")
	}

	// 待接受邀请被拒绝后：允许终审。
	r3.Status = constants.ReviewStatusDeclined
	if err := store.reviews.Update(ctx, r3); err != nil {
		t.Fatalf("decline invite: %v", err)
	}
	updated, err := svc.FinalDecision(ctx, 2, paper.ID, decision)
	if err != nil {
		t.Fatalf("final decision should pass: %v", err)
	}
	if updated.Status != constants.PaperStatusAccepted {
		t.Errorf("expected accepted, got %s", updated.Status)
	}
}

// 已过截止日期的待接受邀请在终审前被清扫为已超期，不再阻塞终审。
func TestPaperServiceFinalDecisionSweepsExpiredInvite(t *testing.T) {
	store, svc := newPaperTestEnv(t)
	ctx := context.Background()

	paper := &model.Paper{
		Title: "超期清扫论文", Status: constants.PaperStatusExternalReview,
		SubmitterID: 1, Version: 1,
	}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}
	past := time.Now().Add(-24 * time.Hour)
	seed := []model.Review{
		{PaperID: paper.ID, ReviewerID: 31, Status: constants.ReviewStatusCompleted, Round: 1},
		{PaperID: paper.ID, ReviewerID: 32, Status: constants.ReviewStatusCompleted, Round: 1},
		{PaperID: paper.ID, ReviewerID: 33, Status: constants.ReviewStatusInvited, Round: 1, DueDate: &past},
	}
	for i := range seed {
		if err := store.reviews.Create(ctx, &seed[i]); err != nil {
			t.Fatalf("create review: %v", err)
		}
	}
	if _, err := svc.FinalDecision(ctx, 2, paper.ID, dto.FinalDecisionRequest{
		Decision: constants.PaperStatusAccepted,
	}); err != nil {
		t.Fatalf("final decision should pass after expiring overdue invite: %v", err)
	}
	expired, _ := store.reviews.FindByID(ctx, seed[2].ID)
	if expired.Status != constants.ReviewStatusExpired {
		t.Errorf("expected overdue invite expired, got %s", expired.Status)
	}
}

func TestPaperServiceSearchLibrary(t *testing.T) {
	store, svc := newPaperTestEnv(t)
	ctx := context.Background()

	for _, st := range []string{constants.PaperStatusAccepted, constants.PaperStatusSubmitted} {
		p := &model.Paper{
			Title: "检索论文" + st, Abstract: "摘要", Keywords: "检索", Subject: "computer",
			Status: st, SubmitterID: 1,
		}
		if err := store.papers.Create(ctx, p); err != nil {
			t.Fatalf("create: %v", err)
		}
	}
	items, total, err := svc.SearchLibrary(ctx, "检索", "", 1, 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 accepted paper, got total=%d len=%d", total, len(items))
	}
	if items[0].Status != constants.PaperStatusAccepted {
		t.Errorf("expected accepted only")
	}
}
