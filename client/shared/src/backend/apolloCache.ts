import { InMemoryCache } from '@apollo/client'

import type { TypedTypePolicies } from '../graphql-operations'

// Defines how the Apollo cache interacts with our GraphQL schema.
// See https://www.apollographql.com/docs/react/caching/cache-configuration/#typepolicy-fields
const typePolicies: TypedTypePolicies = {
    Person: {
        // Replace existing `Person` with the incoming value.
        // Required because of the missing `id` on the `Person` field.
        merge(existing, incoming) {
            return incoming
        },
    },
    Query: {
        fields: {
            node: {
                // Node is a top-level interface field used to easily fetch from different parts of the schema through the relevant `id`.
                // We always want to merge responses from this field as it will be used through very different queries.
                merge: true,
            },
            site: {
                merge: true,
            },
        },
    },
}

export const generateCache = (): InMemoryCache =>
    new InMemoryCache({
        typePolicies,
        possibleTypes: {
            BatchSpecWorkspace: ['VisibleBatchSpecWorkspace', 'HiddenBatchSpecWorkspace'],
            ChangesetSpec: ['VisibleChangesetSpec', 'HiddenChangesetSpec'],
            Changeset: ['ExternalChangeset', 'HiddenExternalChangeset'],
            TeamMember: ['User'],
            Owner: ['Person', 'Team'],
            TreeEntry: ['GitBlob', 'GitTree'],
            File2: ['GitBlob', 'VirtualFile'],
        },
    })

export const cache = generateCache()
