package httpserver_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thienhaole92/uframework/httpserver"
)

func TestResponseWithPaginationSerialization(t *testing.T) {
	t.Parallel()

	resp := httpserver.Response{
		RequestID: "12345",
		Data:      "Test data",
		Pagination: &httpserver.Pagination{
			Limit:     10,
			Total:     100,
			TotalPage: 10,
		},
	}

	jsonData, err := json.Marshal(resp)
	require.NoError(t, err)

	expected := `{"requestId":"12345","data":"Test data","pagination":{"limit":10,"total":100,"totalPage":10}}`

	require.JSONEq(t, expected, string(jsonData))
}

func TestResponseWithoutPaginationSerialization(t *testing.T) {
	t.Parallel()

	resp := httpserver.Response{
		RequestID:  "67890",
		Data:       "No pagination data",
		Pagination: nil,
	}

	jsonData, err := json.Marshal(resp)
	require.NoError(t, err)

	expected := `{"requestId":"67890","data":"No pagination data"}`

	require.JSONEq(t, expected, string(jsonData))
}

func TestPaginationWithValidValuesSerialization(t *testing.T) {
	t.Parallel()

	pag := httpserver.Pagination{
		Limit:     20,
		Total:     200,
		TotalPage: 10,
	}

	jsonData, err := json.Marshal(pag)
	require.NoError(t, err)

	expected := `{"limit":20,"total":200,"totalPage":10}`

	require.JSONEq(t, expected, string(jsonData))
}

func TestPaginationWithZeroValuesSerialization(t *testing.T) {
	t.Parallel()

	pag := httpserver.Pagination{
		Limit:     0,
		Total:     0,
		TotalPage: 0,
	}

	jsonData, err := json.Marshal(pag)
	require.NoError(t, err)

	expected := `{"limit":0,"total":0,"totalPage":0}`

	require.JSONEq(t, expected, string(jsonData))
}
