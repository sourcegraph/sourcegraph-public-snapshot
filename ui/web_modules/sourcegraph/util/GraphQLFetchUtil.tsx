import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { defaultFetch } from "sourcegraph/util/xhr";

const fetch = singleflightFetch(defaultFetch);

export function fetchGraphQLQuery(query: string, variables: Object): Promise<GQL.IQuery> {
	return fetch(`/.api/graphql`, {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
		},
		body: JSON.stringify({
			query: query,
			variables: variables,
		}),
	})
		.then(resp => resp.json())
		.then((resp: GQL.IGraphQLResponseRoot) => {
			if (!resp.data) {
				throw new Error("content not available");
			}
			return Promise.resolve(resp.data);
		});
}
