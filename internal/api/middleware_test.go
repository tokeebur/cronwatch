package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingMiddleware_SetsStatus(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	wrapped := loggingMiddleware(handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestRecoveryMiddleware_CatchesPanic(t *testing.T) {
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	wrapped := recoveryMiddleware(panicking)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)

	// Should not panic the test itself.
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestChain_AppliesMiddlewareInOrder(t *testing.T) {
	var order []string

	makeMiddleware := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+":before")
				next.ServeHTTP(w, r)
				order = append(order, name+":after")
			})
		}
	}

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	h := chain(base, makeMiddleware("A"), makeMiddleware("B"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	want := []string{"A:before", "B:before", "handler", "B:after", "A:after"}
	if len(order) != len(want) {
		t.Fatalf("expected %v, got %v", want, order)
	}
	for i, v := range want {
		if order[i] != v {
			t.Errorf("step %d: expected %q, got %q", i, v, order[i])
		}
	}
}

func TestResponseWriter_DefaultStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := newResponseWriter(rec)
	_, _ = rw.Write([]byte("hello"))

	if rw.status != http.StatusOK {
		t.Fatalf("expected default status 200, got %d", rw.status)
	}
	if rw.bytes != 5 {
		t.Fatalf("expected 5 bytes written, got %d", rw.bytes)
	}
}
