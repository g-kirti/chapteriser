package transcribe

import "testing"

func TestFindHeadings_NumberedChapter(t *testing.T) {
	json := `{"result":[
		{"word":"chapter","start":1.00,"end":1.20},
		{"word":"one","start":1.21,"end":1.35}
	]}`

	chapters := FindHeadings(json, "en")
	if len(chapters) != 1 {
		t.Fatalf("expected 1 chapter, got %d", len(chapters))
	}
	if chapters[0].Title != "chapter one" {
		t.Fatalf("unexpected phrase: %q", chapters[0].Title)
	}
}

func TestFindHeadings_DoesNotAppendFollowingSentenceWord(t *testing.T) {
	json := `{"result":[
		{"word":"chapter","start":10.00,"end":10.15},
		{"word":"two","start":10.20,"end":10.35},
		{"word":"three","start":11.30,"end":11.45}
	]}`

	chapters := FindHeadings(json, "en")
	if len(chapters) != 1 {
		t.Fatalf("expected 1 chapter, got %d", len(chapters))
	}
	if chapters[0].Title != "chapter two" {
		t.Fatalf("expected %q, got %q", "chapter two", chapters[0].Title)
	}
}

func TestFindHeadings_StandaloneHeadingNeedsBoundaryOrGap(t *testing.T) {
	json := `{"result":[
		{"word":"introduction","start":100.00,"end":100.10},
		{"word":"of","start":100.12,"end":100.20},
		{"word":"chapter","start":150.00,"end":150.12},
		{"word":"one","start":150.14,"end":150.30},
		{"word":"epilogue","start":400.00,"end":400.20}
	]}`

	chapters := FindHeadings(json, "en")
	if len(chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(chapters))
	}
	if chapters[0].Title != "chapter one" {
		t.Fatalf("expected first chapter %q, got %q", "chapter one", chapters[0].Title)
	}
	if chapters[1].Title != "epilogue" {
		t.Fatalf("expected second chapter %q, got %q", "epilogue", chapters[1].Title)
	}
}
