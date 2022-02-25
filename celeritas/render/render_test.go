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
	// GO TEMPLATE FILE WAS FOUND
	err = testRenderer.Page(w, r, "home", nil, nil)
	if err != nil {
		t.Error("Error rendering Go template page", err)
	}
	// GO TEMPLATE DOESN'T EXIST
	err = testRenderer.Page(w, r, "no-go-file", nil, nil)
	if err == nil {
		t.Error("Error rendering non-existent Go template", err)
	}

	////////////////////////////////////////////////
	// TEST JET TEMPLATES
	////////////////////////////////////////////////
	testRenderer.Renderer = "jet"
	// JET TEMPLATE FILE WAS FOUND
	err = testRenderer.Page(w, r, "home", nil, nil)
	if err != nil {
		t.Error("Error rendering Jet template page", err)
	}
	// JET TEMPLATE DOESN'T EXIST
	err = testRenderer.Page(w, r, "no-jet-file", nil, nil)
	if err == nil {
		t.Error("Error rendering non-existent Jet template", err)
	}

	////////////////////////////////////////////////
	// NO TEMPLATE ENGINE SPECIFIED
	////////////////////////////////////////////////
	testRenderer.Renderer = ""
	err = testRenderer.Page(w, r, "home", nil, nil)
	if err == nil {
		t.Error("No error returned when template engine is not specified", err)
	}

}

func TestRender_GoPage(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		t.Error(err)
	}

	testRenderer.Renderer = "go"
	// GO TEMPLATE FILE WAS FOUND
	err = testRenderer.Page(w, r, "home", nil, nil)
	if err != nil {
		t.Error("Error rendering Go template page", err)
	}
	// GO TEMPLATE DOESN'T EXIST
	err = testRenderer.Page(w, r, "no-go-file", nil, nil)
	if err == nil {
		t.Error("Error rendering non-existent Go template", err)
	}
}

func TestRender_JetPage(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		t.Error(err)
	}

	testRenderer.Renderer = "jet"
	// JET TEMPLATE FILE WAS FOUND
	err = testRenderer.Page(w, r, "home", nil, nil)
	if err != nil {
		t.Error("Error rendering Jet template page", err)
	}
	// GO TEMPLATE DOESN'T EXIST
	err = testRenderer.Page(w, r, "no-jet-file", nil, nil)
	if err == nil {
		t.Error("Error rendering non-existent Jet template", err)
	}
}
