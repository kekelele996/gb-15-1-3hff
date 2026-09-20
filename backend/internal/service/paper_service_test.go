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

func TestPaperServiceReviseStartsNewRound(t *testing.T) {
	store, svc := newPaperTestEnv(t)
	ctx := context.Background()

	paper := &model.Paper{
		Title: "多轮评审论文", Abstract: "摘要内容长度满足要求。", Keywords: "评审", Subject: "computer",
		AuthorsMeta: `[{"name":"张三"}]`, Status: constants.PaperStatusRevision, Version: 1, SubmitterID: 1,
	}
	if err := store.papers.Create(ctx, paper); err != nil {
		t.Fatalf("create paper: %v", err)
	}
	future := time.Now().Add(7 * 24 * time.Hour)
	past := time.Now().Add(-time.Hour)
	seed := []model.Review{
		{PaperID: paper.ID, ReviewerID: 11, Status: constants.ReviewStatusCompleted, Round: 1, Decision: constants.ReviewDecisionMinorRevision},
		{PaperID: paper.ID, ReviewerID: 12, Status: constants.ReviewStatusAccepted, Round: 1, DueDate: &future},
		{PaperID: paper.ID, ReviewerID: 13, Status: constants.ReviewStatusDeclined, Round: 1},
		{PaperID: paper.ID, ReviewerID: 14, Status: constants.ReviewStatusInvited, Round: 1, DueDate: &past},
	}
	for i := range seed {
		if err := store.reviews.Create(ctx, &seed[i]); err != nil {
			t.Fatalf("seed review: %v", err)
		}
	}

	if _, err := svc.Revise(ctx, 1, paper.ID, dto.ReviseRequest{
		FileKey: "papers/v2.pdf", FileName: "v2.pdf", ResponseLetter: "已逐条回复审稿意见并完成修改。",
	}); err != nil {
		t.Fatalf("revise: %v", err)
	}

	reviews, err := store.reviews.ListByPaper(ctx, paper.ID)
	if err != nil {
		t.Fatalf("list reviews: %v", err)
	}
	round2 := map[uint]model.Review{}
	for _, r := range reviews {
		if r.Round == 2 {
			round2[r.ReviewerID] = r
		}
	}
	// 仍在有效期内的审稿人（已完成 + 已接受未超期）开启新一轮匿名评审
	if len(round2) != 2 {
		t.Fatalf("expected 2 round-2 tasks, got %d", len(round2))
	}
	for _, id := range []uint{11, 12} {
		r, ok := round2[id]
		if !ok {
			t.Fatalf("expected round-2 task for reviewer %d", id)
		}
		if r.Status != constants.ReviewStatusInvited || r.DueDate == nil {
			t.Errorf("unexpected round-2 task: %+v", r)
		}
	}
	// 旧开放任务失效，旧意见只读保留
	old1, _ := store.reviews.FindByID(ctx, seed[0].ID)
	if old1.Status != constants.ReviewStatusCompleted || old1.Decision != constants.ReviewDecisionMinorRevision {
		t.Errorf("old completed opinion must stay read-only, got %+v", old1)
	}
	old2, _ := store.reviews.FindByID(ctx, seed[1].ID)
	if old2.Status != constants.ReviewStatusExpired {
		t.Errorf("old open task must be expired, got %s", old2.Status)
	}
	old4, _ := store.reviews.FindByID(ctx, seed[3].ID)
	if old4.Status != constants.ReviewStatusExpired {
		t.Errorf("overdue invite must be expired, got %s", old4.Status)
	}
}

func TestPaperServiceFinalDecisionGate(t *testing.T) {
	newPaperWithReviews := func(t *testing.T, store *fakeStore, reviews []model.Review) *model.Paper {
		t.Helper()
		paper := &model.Paper{
			Title: "终审论文", Abstract: "摘要内容长度满足要求。", Keywords: "终审", Subject: "computer",
			AuthorsMeta: `[{"name":"张三"}]`, Status: constants.PaperStatusExternalReview, Version: 1, SubmitterID: 1,
		}
		if err := store.papers.Create(context.Background(), paper); err != nil {
			t.Fatalf("create paper: %v", err)
		}
		for i := range reviews {
			reviews[i].PaperID = paper.ID
			if err := store.reviews.Create(context.Background(), &reviews[i]); err != nil {
				t.Fatalf("seed review: %v", err)
			}
		}
		return paper
	}
	future := time.Now().Add(7 * 24 * time.Hour)
	past := time.Now().Add(-time.Hour)

	t.Run("blocked when less than two completed", func(t *testing.T) {
		store, svc := newPaperTestEnv(t)
		paper := newPaperWithReviews(t, store, []model.Review{
			{ReviewerID: 11, Status: constants.ReviewStatusCompleted, Round: 1},
			{ReviewerID: 12, Status: constants.ReviewStatusAccepted, Round: 1, DueDate: &future},
		})
		_, err := svc.FinalDecision(context.Background(), 2, paper.ID, dto.FinalDecisionRequest{Decision: constants.PaperStatusAccepted})
		if err == nil {
			t.Fatalf("expected gate error with only 1 completed review")
		}
	})

	t.Run("blocked when pending invite exists", func(t *testing.T) {
		store, svc := newPaperTestEnv(t)
		paper := newPaperWithReviews(t, store, []model.Review{
			{ReviewerID: 11, Status: constants.ReviewStatusCompleted, Round: 1},
			{ReviewerID: 12, Status: constants.ReviewStatusCompleted, Round: 1},
			{ReviewerID: 13, Status: constants.ReviewStatusInvited, Round: 1, DueDate: &future},
		})
		_, err := svc.FinalDecision(context.Background(), 2, paper.ID, dto.FinalDecisionRequest{Decision: constants.PaperStatusAccepted})
		if err == nil {
			t.Fatalf("expected gate error with pending invite")
		}
	})

	t.Run("allowed with two completed and overdue invite swept", func(t *testing.T) {
		store, svc := newPaperTestEnv(t)
		paper := newPaperWithReviews(t, store, []model.Review{
			{ReviewerID: 11, Status: constants.ReviewStatusCompleted, Round: 1},
			{ReviewerID: 12, Status: constants.ReviewStatusCompleted, Round: 1},
			{ReviewerID: 13, Status: constants.ReviewStatusInvited, Round: 1, DueDate: &past},
		})
		updated, err := svc.FinalDecision(context.Background(), 2, paper.ID, dto.FinalDecisionRequest{Decision: constants.PaperStatusAccepted})
		if err != nil {
			t.Fatalf("final decision: %v", err)
		}
		if updated.Status != constants.PaperStatusAccepted {
			t.Errorf("expected accepted, got %s", updated.Status)
		}
	})
}
