// We want to limit the number of imported modules as much as possible
/* eslint-disable no-restricted-imports */
export {
    getDocumentNode,
    type GraphQLClient,
    gql,
    dataOrThrowErrors,
    type GraphQLResult,
} from '@sourcegraph/http-client/src/graphql/graphql'
export { fromObservableQuery } from '@sourcegraph/http-client/src/graphql/apollo/fromObservableQuery'
