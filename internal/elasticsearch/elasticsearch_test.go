// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package elasticsearch_test

import (
	"testing"

	"github.com/juju/charmstore/internal/elasticsearch"
	"github.com/juju/charmstore/internal/storetesting"
	jujutesting "github.com/juju/testing"
	gc "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) {
	storetesting.ElasticSearchTestPackage(t, nil)
}

type Suite struct {
	jujutesting.IsolationSuite
	storetesting.ElasticSearchSuite
}

func (s *Suite) SetUpSuite(c *gc.C) {
	s.IsolationSuite.SetUpSuite(c)
	s.ElasticSearchSuite.SetUpSuite(c)
}
func (s *Suite) TearDownSuite(c *gc.C) {
	s.ElasticSearchSuite.TearDownSuite(c)
	s.IsolationSuite.TearDownSuite(c)
}
func (s *Suite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	s.ElasticSearchSuite.SetUpTest(c)
	s.NewIndex(c)
}
func (s *Suite) TearDownTest(c *gc.C) {
	s.ElasticSearchSuite.TearDownTest(c)
	s.IsolationSuite.TearDownTest(c)
}

var _ = gc.Suite(&Suite{})

func (s *Suite) TestSuccessfulPostDocument(c *gc.C) {
	doc := map[string]string{
		"a": "b",
	}
	id, err := s.ES.PostDocument(s.TestIndex, "testtype", doc)
	c.Assert(err, gc.IsNil)
	c.Assert(id, gc.NotNil)
	var result map[string]string
	err = s.ES.GetDocument(s.TestIndex, "testtype", id, &result)
	c.Assert(err, gc.IsNil)
}

func (s *Suite) TestSuccessfulPutNewDocument(c *gc.C) {
	doc := map[string]string{
		"a": "b",
	}
	// Show that no document with this id exists.
	exists, err := s.ES.EnsureID(s.TestIndex, "testtype", "a")
	c.Assert(err, gc.IsNil)
	c.Assert(exists, gc.Equals, false)
	err = s.ES.PutDocument(s.TestIndex, "testtype", "a", doc)
	c.Assert(err, gc.IsNil)
	var result map[string]string
	err = s.ES.GetDocument(s.TestIndex, "testtype", "a", &result)
	c.Assert(result["a"], gc.Equals, "b")
	exists, err = s.ES.EnsureID(s.TestIndex, "testtype", "a")
	c.Assert(err, gc.IsNil)
	c.Assert(exists, gc.Equals, true)

}

func (s *Suite) TestSuccessfulPutUpdatedDocument(c *gc.C) {
	doc := map[string]string{
		"a": "b",
	}
	err := s.ES.PutDocument(s.TestIndex, "testtype", "a", doc)
	c.Assert(err, gc.IsNil)
	doc["a"] = "c"
	err = s.ES.PutDocument(s.TestIndex, "testtype", "a", doc)
	c.Assert(err, gc.IsNil)
	var result map[string]string
	err = s.ES.GetDocument(s.TestIndex, "testtype", "a", &result)
	c.Assert(result["a"], gc.Equals, "c")
}

func (s *Suite) TestDelete(c *gc.C) {
	doc := map[string]string{
		"a": "b",
	}
	_, err := s.ES.PostDocument(s.TestIndex, "testtype", doc)
	c.Assert(err, gc.IsNil)
	err = s.ES.DeleteIndex(s.TestIndex)
	c.Assert(err, gc.IsNil)
}

func (s *Suite) TestDeleteErrorOnNonExistingIndex(c *gc.C) {
	err := s.ES.DeleteIndex("nope")
	terr := err.(*elasticsearch.ErrNotFound)
	c.Assert(terr.Message(), gc.Equals, "index nope not found")
}

func (s *Suite) TestIndexesCreatedAutomatically(c *gc.C) {
	doc := map[string]string{"a": "b"}
	_, err := s.ES.PostDocument(s.TestIndex, "testtype", doc)
	c.Assert(err, gc.IsNil)
	indexes, err := s.ES.ListAllIndexes()
	c.Assert(err, gc.IsNil)
	c.Assert(indexes, gc.Not(gc.HasLen), 0)
	found := false
	for _, index2 := range indexes {
		if index2 == s.TestIndex {
			found = true
		}
	}
	c.Assert(found, gc.Equals, true)
}

func (s *Suite) TestSearch(c *gc.C) {
	doc := map[string]string{"foo": "bar"}
	_, err := s.ES.PostDocument(s.TestIndex, "testtype", doc)
	c.Assert(err, gc.IsNil)
	doc["foo"] = "baz"
	id2, err := s.ES.PostDocument(s.TestIndex, "testtype", doc)
	c.Assert(err, gc.IsNil)
	s.ES.RefreshIndex(s.TestIndex)
	query := `{"query": {"term": {"foo": "baz"}}}`
	results, err := s.ES.Search(s.TestIndex, "testtype", query)
	c.Assert(err, gc.IsNil)
	c.Assert(results.Hits.Total, gc.Equals, 1)
	c.Assert(results.Hits.Hits[0].ID, gc.Equals, id2)
}
