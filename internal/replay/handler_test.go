package replay_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/retryq/internal/deadletter"
	"github.com/your-org/retryq/internal/replay"
)

type fakeStore struct{ entries []deadletter.Entry }

func (f *fakeStore) List() []deadletter.Entry { return f.entries }

func handlerOpts() replay.Options {
	opts := replay.DefaultOptions()
	opts.Logger = silentLog
	return opts
}

func TestHandler_WrongMethod(t *testing.T) {
	h := replay.NewHandler(&fakeStore{}, &stubDispatcher{}, handlerOpts())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/replay", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_EmptyStore(t *testing.T) {
	h := replay.NewHandler(&fakeStore{}, &stubDispatcher{}, handlerOpts())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/replay", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp struct {
		Total int `json:"total"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Total != 0 {
		t.Errorf("expected total=0, got %d", resp.Total)
	}
}

func TestHandler_AllSucceed(t *testing.T) {
	store := &fakeStore{entries: []deadletter.Entry{makeEntry("1"), makeEntry("2")}}
	h := replay.NewHandler(store, &stubDispatcher{}, handlerOpts())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/replay", nil))

	var resp struct {
		Total     int `json:"total"`
		Succeeded int `json:"succeeded"`
		Failed    int `json:"failed"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Total != 2 || resp.Succeeded != 2 || resp.Failed != 0 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestHandler_PartialFailure(t *testing.T) {
	store := &fakeStore{entries: []deadletter.Entry{makeEntry("a"), makeEntry("b")}}
	errDisp := &errorDispatcher{failID: "a", err: errors.New("boom")}
	h := replay.NewHandler(store, errDisp, handlerOpts())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/replay", nil))

	var resp struct {
		Succeeded int `json:"succeeded"`
		Failed    int `json:"failed"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Succeeded != 1 || resp.Failed != 1 {
		t.Errorf("expected 1 succeeded 1 failed, got %+v", resp)
	}
}

type errorDispatcher struct {
	failID string
	err    error
}

func (e *errorDispatcher) Dispatch(_ context.Context, entry deadletter.Entry) error {
	if entry.ID == e.failID {
		return e.err
	}
	return nil
}
