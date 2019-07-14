package graphql

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

func TestSubscription(t *testing.T) {
	var maxPublish = 5
	m := make(chan interface{})

	source1 := source.NewSource(&source.Source{
		Body: []byte(`subscription {
			watch_count
		}`),
		Name: "GraphQL request",
	})

	source2 := source.NewSource(&source.Source{
		Body: []byte(`subscription {
			watch_should_fail
		}`),
		Name: "GraphQL request",
	})

	document1, _ := parser.Parse(parser.ParseParams{Source: source1})
	document2, _ := parser.Parse(parser.ParseParams{Source: source2})

	schema, err := NewSchema(SchemaConfig{
		Query: NewObject(ObjectConfig{
			Name: "Query",
			Fields: Fields{
				"hello": &Field{
					Type: String,
					Resolve: func(p ResolveParams) (interface{}, error) {
						return "world", nil
					},
				},
			},
		}),
		Subscription: NewObject(ObjectConfig{
			Name: "Subscription",
			Fields: Fields{
				"watch_count": &Field{
					Type: String,
					Resolve: func(p ResolveParams) (interface{}, error) {
						return fmt.Sprintf("count=%v", p.Source), nil
					},
					Subscribe: func(p ResolveParams) (interface{}, error) {
						return m, nil
					},
				},
				"watch_should_fail": &Field{
					Type: String,
					Resolve: func(p ResolveParams) (interface{}, error) {
						return fmt.Sprintf("count=%v", p.Source), nil
					},
					Subscribe: func(p ResolveParams) (interface{}, error) {
						return nil, nil
					},
				},
			},
		}),
	})

	if err != nil {
		t.Errorf("failed to create schema: %v", err)
		return
	}

	failIterator := Subscribe(SubscribeParams{
		Schema:   schema,
		Document: document2,
	})

	// test a subscribe that should fail due to no return value
	failIterator.ForEach(func(count int64, res *Result, doneFunc func()) {
		if !res.HasErrors() {
			t.Errorf("subscribe failed to catch nil result from subscribe")
			doneFunc()
			return
		}
		doneFunc()
		return
	})

	resultIterator := Subscribe(SubscribeParams{
		Schema:       schema,
		Document:     document1,
		ContextValue: context.Background(),
	})

	resultIterator.ForEach(func(count int64, res *Result, doneFunc func()) {
		if res.HasErrors() {
			t.Errorf("subscribe error(s): %v", res.Errors)
			doneFunc()
			return
		}

		if res.Data != nil {
			data := res.Data.(map[string]interface{})["watch_count"]
			expected := fmt.Sprintf("count=%d", count)
			actual := fmt.Sprintf("%v", data)
			if actual != expected {
				t.Errorf("subscription result error: expected %q, actual %q", expected, actual)
				doneFunc()
				return
			}

			// test the done func by quitting after 3 iterations
			// the publisher will publish up to 5
			if count >= int64(maxPublish-2) {
				doneFunc()
				return
			}
		}
	})

	// start publishing
	go func() {
		for i := 1; i <= maxPublish; i++ {
			time.Sleep(200 * time.Millisecond)
			m <- i
		}
	}()

	// give time for the test to complete
	time.Sleep(1 * time.Second)
}