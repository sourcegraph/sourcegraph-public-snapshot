package httpapi

import (
	"encoding/json"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

func serveGraphQL(w http.ResponseWriter, r *http.Request) (err error) {
	if r.Method == "GET" {
		w.Header().Set("content-type", "text/html")
		w.Write(graphiqlPage)
		return nil
	}

	var params struct {
		Query     string `json:"query"`
		Variables map[string]interface{}
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}

	response := backend.GraphQLSchema.Exec(r.Context(), params.Query, "", params.Variables)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	w.Write(responseJSON)

	return nil
}

var graphiqlPage = []byte(`
<!DOCTYPE html>
<html>
	<head>
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.7.8/graphiql.css" />
		<script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/1.0.0/fetch.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.3.2/react.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.3.2/react-dom.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.7.8/graphiql.js"></script>
	</head>
	<body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
		<div id="graphiql" style="height: 100vh;">Loading...</div>
		<script>
			function graphQLFetcher(graphQLParams) {
				graphQLParams.variables = graphQLParams.variables ? JSON.parse(graphQLParams.variables) : null;
				return fetch("/.api/graphql", {
					method: "post",
					body: JSON.stringify(graphQLParams),
					credentials: "include",
				}).then(function (response) {
					return response.text();
				}).then(function (responseBody) {
					try {
						return JSON.parse(responseBody);
					} catch (error) {
						return responseBody;
					}
				});
			}

			ReactDOM.render(
				React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
				document.getElementById("graphiql")
			);
		</script>
	</body>
</html>
`)
