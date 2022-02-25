package render

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRender_Page(t *testing.T) {
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		t.Error(err)
	}

	w := httptest.NewRecorder()

	testRenderer.RootPath = "./testdata"

	////////////////////////////////////////////////
	// TEST GO TEMPLATES
	////////////////////////////////////////////////
	testRenderer.Renderer = "go"
	err = testRenderer.Page(w, r, "home", nil, nil)
	if err != nil {
		t.Error("Error rendering page", err)
	}

	////////////////////////////////////////////////
	// TEST JET TEMPLATES
	////////////////////////////////////////////////
	testRenderer.Renderer = "jet"
	err = testRenderer.Page(w, r, "home", nil, nil)
	if err != nil {
		t.Error("Error rendering page", err)
	}

}
