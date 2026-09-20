package service

import (
	"context"
	"testing"

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
