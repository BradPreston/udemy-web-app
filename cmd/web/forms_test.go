package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Has(t *testing.T) {
	form := NewForm(nil)
	has := form.Has("whatever")
	if has == true {
		t.Error("form shows it has a field when it should not")
	}
	postedData := url.Values{}
	postedData.Add("someKey", "someValue")
	form = NewForm(postedData)
	has = form.Has("someKey")
	if has == false {
		t.Error("form shows it does not have a field when it should")
	}
}

func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POST", "/test", nil)
	form := NewForm(r.PostForm)
	form.Required("a", "b", "c")
	if form.Valid() == true {
		t.Error("form shows valid when missing required fields")
	}
	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")
	r, _ = http.NewRequest("POST", "/test", nil)
	r.PostForm = postedData
	form = NewForm(r.PostForm)
	form.Required("a", "b", "c")
	if form.Valid() == false {
		t.Error("shows post does not have required fields when it does")
	}
}

func TestForm_Check(t *testing.T) {
	form := NewForm(nil)
	form.Check(false, "password", "password is required")
	if form.Valid() == true {
		t.Error("valid() returns false and it should be true when calling check()")
	}
}

func TestForm_Error_Get(t *testing.T) {
	form := NewForm(nil)
	form.Check(false, "password", "password is required")
	s := form.Errors.Get("password")
	if len(s) == 0 {
		t.Error("should have an error returned from Get() but do not")
	}
	s = form.Errors.Get("email")
	if len(s) != 0 {
		t.Error("should not have an error, but got one from Get()")
	}
}
