// We want to limit the number of imported modules as much as possible
/* eslint-disable no-restricted-imports */
export { fromObservableQuery } from '@sourcegraph/http-client/src/graphql/apollo/fromObservableQuery'
export {
    dataOrThrowErrors,
    getDocumentNode,
    gql,
    type GraphQLClient,
    type GraphQLResult,
} from '@sourcegraph/http-client/src/graphql/graphql'
