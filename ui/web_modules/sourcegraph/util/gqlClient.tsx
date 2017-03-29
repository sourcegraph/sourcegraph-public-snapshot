import { ApolloClient, ApolloQueryResult, createNetworkInterface } from "apollo-client";
import gql from "graphql-tag";

import { context } from "sourcegraph/app/context";

export const gqlClient = new ApolloClient({
	networkInterface: createNetworkInterface({
		uri: `${context.appURL}/.api/graphql`,
		opts: {
			headers: context.xhrHeaders,
			credentials: "include",
		},
	}),
});

export function fetchGQL(query: string, variables?: Object): Promise<ApolloQueryResult<GQL.IQuery>> {
	return gqlClient.query<GQL.IQuery>({
		query: gql`${query}`,
		variables
	});
}
