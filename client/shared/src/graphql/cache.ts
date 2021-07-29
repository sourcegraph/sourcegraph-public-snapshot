import { InMemoryCache } from '@apollo/client'

import { TypedTypePolicies } from '../graphql-operations'

// Defines how the Apollo cache interacts with our GraphQL schema.
// See https://www.apollographql.com/docs/react/caching/cache-configuration/#typepolicy-fields
const typePolicies: TypedTypePolicies = {}

export const cache = new InMemoryCache({
    typePolicies,
})
