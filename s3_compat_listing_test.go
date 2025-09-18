package eightfs_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/8fs-io/core/internal/transport/http/handlers"
	"github.com/stretchr/testify/assert"
)

// This test focuses on delimiter/common prefixes and marker continuation semantics.
func TestS3_List_Delimiter_CommonPrefixes_And_Marker(t *testing.T) {
	r, cfg := newTestRouter(t, nil)
	key := cfg.Auth.DefaultKey.AccessKey

	// Create bucket
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/delim-bkt", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Upload objects across two top-level prefixes
	files := []string{
		"foo/a.txt",
		"foo/b.txt",
		"bar/c.txt",
		"bar/d.txt",
	}
	for _, k := range files {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("PUT", "/delim-bkt/"+k, strings.NewReader("x"))
		req.Header.Set("Authorization", authHeader(key))
		req.Header.Set("Content-Type", "text/plain")
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	}

	// Delimiter at root -> expect both contents and common prefixes (implementation returns objects not grouped by delimiter)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/delim-bkt?delimiter=/", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var list handlers.ListBucketResult
	parseXML(t, w.Body.Bytes(), &list)
	// With a delimiter, objects that are rolled up into a common prefix should not be returned, so Contents should be empty.
	assert.Empty(t, list.Contents)
	var cps []string
	for _, p := range list.CommonPrefixes {
		cps = append(cps, p.Prefix)
	}
	assert.ElementsMatch(t, []string{"bar/", "foo/"}, cps)

	// Delimiter with prefix=foo/ -> should return only objects under that prefix not matching the delimiter.
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/delim-bkt?prefix=foo/&delimiter=/", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	list = handlers.ListBucketResult{}
	parseXML(t, w.Body.Bytes(), &list)
	var returned []string
	for _, c := range list.Contents {
		returned = append(returned, c.Key)
	}
	assert.ElementsMatch(t, []string{"foo/a.txt", "foo/b.txt"}, returned)
	// CommonPrefixes may be empty when scoped to prefix without deeper subdirs

	// Pagination with max-keys=1 to force truncation and continuation
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/delim-bkt?max-keys=1", nil)
	req.Header.Set("Authorization", authHeader(key))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	list = handlers.ListBucketResult{}
	parseXML(t, w.Body.Bytes(), &list)
	assert.Len(t, list.Contents, 1)
	var next string
	if list.IsTruncated {
		assert.NotEmpty(t, list.NextMarker)
		next = list.NextMarker
	}

	if next != "" {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/delim-bkt?marker="+next, nil)
		req.Header.Set("Authorization", authHeader(key))
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		list = handlers.ListBucketResult{}
		parseXML(t, w.Body.Bytes(), &list)
		for _, c := range list.Contents {
			assert.Greater(t, c.Key, next) // service filters strictly > marker
		}
	}
}