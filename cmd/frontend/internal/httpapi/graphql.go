package httpapi

import (
	"net/http"

	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/graphqlbackend"
)

var relayHandler = &relay.Handler{Schema: graphqlbackend.GraphQLSchema}

func serveGraphQL(w http.ResponseWriter, r *http.Request) (err error) {
	if r.Method == "GET" {
		// Allow displaying in an iframe on the /api/console page.
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")

		w.Header().Set("Content-Type", "text/html")
		w.Write(graphiqlPage)
		return nil
	}
	relayHandler.ServeHTTP(w, r)
	return nil
}

var graphiqlPage = []byte(`
<!DOCTYPE html>
<html>
	<head>
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.css" />
		<script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/1.0.0/fetch.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.4.2/react.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.4.2/react-dom.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.js"></script>
		<style>.graphiql-container .editorWrap { overflow: hidden; }</style> <!-- hide vertical scrollbars for ~2px -->
	</head>
	<body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
		<div id="graphiql" style="height: 100vh;">Loading...</div>
		<script>
			// URL handling taken from https://github.com/graphql/graphiql/blob/master/example/index.html.

			// Parse the search string to get url parameters.
			// window.location.hash starts with a '#' character.
			var parameters = JSON.parse(decodeURIComponent(window.location.hash.slice(1)) || '{}');
			// if variables was provided, try to format it.
			if (parameters.variables) {
			  try {
				parameters.variables =
				  JSON.stringify(JSON.parse(parameters.variables), null, 2);
			  } catch (e) {
				// Do nothing, we want to display the invalid JSON as a string, rather
				// than present an error.
			  }
			}
			if (Object.keys(parameters).length > 0) {
				sendParametersToParent()
			}

			// When the query and variables string is edited, update the URL bar so
			// that it can be easily shared
			function onEditQuery(newQuery) {
			  parameters.query = newQuery;
			  sendParametersToParent();
			}
			function onEditVariables(newVariables) {
			  parameters.variables = newVariables;
			  sendParametersToParent();
			}
			function onEditOperationName(newOperationName) {
			  parameters.operationName = newOperationName;
			  sendParametersToParent();
			}

			function sendParametersToParent() {
				window.parent.postMessage(parameters, window.location.origin)
			}

			function graphQLFetcher(graphQLParams) {
				return fetch("/.api/graphql", {
					method: "post",
					body: JSON.stringify(graphQLParams),
					credentials: "include",
					headers: new Headers({ "x-requested-with": "Sourcegraph GraphQL Explorer" }), // enables authenticated queries
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

			var node = document.getElementById("graphiql");

			ReactDOM.render(
				React.createElement(GraphiQL, {
					query: parameters.query,
					variables: parameters.variables,
					operationName: parameters.operationName,
					onEditQuery: onEditQuery,
					onEditVariables: onEditVariables,
					onEditOperationName: onEditOperationName,
					fetcher: graphQLFetcher,
					defaultQuery: "# Type queries here, with completion, validation, and hovers.\n#\n# Here's an example query to get you started:\n\nquery {\n  currentUser {\n    username\n  }\n  site {\n    repositories(first: 1) {\n      nodes {\n        uri\n      }\n    }\n  }\n}\n",
				}),
				node
			);

			// Necessary to persist GraphiQL localStorage state upon parent iframe react-router
			// route changes.
			window.onunload = function() {
				ReactDOM.unmountComponentAtNode(node)
			}
		</script>
	</body>
</html>
`)
