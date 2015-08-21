package gql

import (
	"reflect"
	"testing"

	"github.com/chris-ramon/graphql-go/types"

	"./testutil"
)

type T struct {
	Query    string
	Schema   types.GraphQLSchema
	Expected interface{}
}

var (
	Tests = []T{
		T{
			Query: `
				query HeroNameQuery {
					hero {
						name
					}
				}
			`,
			Schema: testutil.StarWarsSchema,
			Expected: map[string]interface{}{
				"name": "R2-D2",
			},
		},
		T{
			Query: `
				query HeroNameAndFriendsQuery {
					hero {
						id
						name
						friends {
							name
						}
					}
				}
				`,
			Schema: testutil.StarWarsSchema,
			Expected: map[string]interface{}{
				"id":   "2001",
				"name": "R2-D2",
				"friends": []map[string]interface{}{
					map[string]interface{}{"name": "Luke Skywalker"},
					map[string]interface{}{"name": "Han Solo"},
					map[string]interface{}{"name": "Leia Organa"},
				},
			},
		},
	}
)

func TestQuery(t *testing.T) {
	for _, test := range Tests {
		graphqlParams := GraphqlParams{
			Schema:        test.Schema,
			RequestString: test.Query,
		}
		testGraphql(test, graphqlParams, t)
	}
}

func testGraphql(test T, p GraphqlParams, t *testing.T) {
	resultChannel := make(chan *types.GraphQLResult)
	go Graphql(p, resultChannel)
	result := <-resultChannel
	if len(result.Errors) > 0 {
		t.Fatalf("wrong result, unexpected errors: %v", result.Errors)
	}
	if !reflect.DeepEqual(result.Data, test.Expected) {
		t.Fatalf("wrong result, query: %v, graphql result: %v, expected: %v", test.Query, result, test.Expected)
	}
}