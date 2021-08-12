import { InMemoryCache } from '@apollo/client'

import { TypedTypePolicies } from '../graphql-operations'

// Defines how the Apollo cache interacts with our GraphQL schema.
// See https://www.apollographql.com/docs/react/caching/cache-configuration/#typepolicy-fields
const typePolicies: TypedTypePolicies = {
    Query: {
        fields: {
            node: {
                // `Node` is a heavily used, top-level query. We always want to merge responses into the existing cache.
                merge: true,
            },
        },
    },
}

export const cache = new InMemoryCache({
    typePolicies,
})
