package graphql

import (
	"context"
	"strings"

	"github.com/graphql-go/graphql/language/ast"

	"github.com/spaceuptech/space-cloud/model"
	"github.com/spaceuptech/space-cloud/utils"
)

func (graph *Module) execUpdateRequest(field *ast.Field, store utils.M) (utils.M, error) {
	dbType := field.Directives[0].Name.Value
	col := strings.TrimPrefix(field.Name.Value, "update_")

	req, err := generateUpdateRequest(field, store)
	if err != nil {
		return nil, err
	}
	status, err := graph.auth.IsUpdateOpAuthorised(graph.project, dbType, col, "", req)
	if err != nil {
		return nil, err
	}

	return utils.M{"status": status}, graph.crud.Update(context.TODO(), dbType, graph.project, col, req)
}

func extractUpdateOperation(args []*ast.Argument, store utils.M) (string, error) {
	for _, v := range args {
		switch v.Name.Value {
		case "op":
			temp, err := ParseValue(v.Value, store)
			if err != nil {
				return "", err
			}
			if temp.(string) == "upsert" {
				return utils.Upsert, nil
			}

			return utils.All, nil
		}
	}
	return utils.All, nil
}

func generateUpdateRequest(field *ast.Field, store utils.M) (*model.UpdateRequest, error) {
	var err error
	var updateRequest model.UpdateRequest

	updateRequest.Operation, err = extractUpdateOperation(field.Arguments, store)
	if err != nil {
		return nil, err
	}

	updateRequest.Find, err = extractWhereClause(field.Arguments, store)
	if err != nil {
		return nil, err
	}

	updateRequest.Update, err = extractUpdateArgs(field.Arguments, store)
	if err != nil {
		return nil, err
	}

	return &updateRequest, nil
}

func extractUpdateArgs(args []*ast.Argument, store utils.M) (utils.M, error) {
	var t map[string]interface{}
	for _, v := range args {
		switch v.Name.Value {
		case "set", "inc", "mul", "max", "min", "currentTimestamp", "currentDate", "push", "rename", "remove":
			temp, err := ParseValue(v.Value, store)
			if err != nil {
				return nil, err
			}
			t["$"+v.Name.Value] = temp
		}
	}
	return t, nil
}
