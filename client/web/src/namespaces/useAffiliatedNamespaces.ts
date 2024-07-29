import { gql, useQuery } from '@sourcegraph/http-client'

import type {
    Scalars,
    ViewerAffiliatedNamespacesResult,
    ViewerAffiliatedNamespacesVariables,
} from '../graphql-operations'

type Namespace = ViewerAffiliatedNamespacesResult['viewer']['affiliatedNamespaces']['nodes'][number]

/**
 * React hook that fetches all affiliated namespaces for the viewer. A user's affiliated namespaces
 * are their own user account plus all organizations of which they are a member. For anonymous
 * visitors (on instances that support anonymous usage), the {@link namespaces} list may be empty
 * and {@link initialNamespace} may be undefined.
 * @param initialNamespaceID The ID of the namespace to return in {@link initialNamespace}. If not
 * provided, the user's user account is used.
 */
export const useAffiliatedNamespaces = (
    initialNamespaceID?: Scalars['ID']
): {
    namespaces?: Namespace[]
    initialNamespace?: Namespace
    loading: boolean
    error?: Error
} => {
    const { data, loading, error } = useQuery<ViewerAffiliatedNamespacesResult, ViewerAffiliatedNamespacesVariables>(
        viewerAffiliatedNamespacesQuery,
        { fetchPolicy: 'cache-and-network' }
    )

    if (error) {
        return { error, loading }
    }
    if (!data || loading) {
        return {
            error: loading ? undefined : new Error('Unable to get affiliated namespaces'),
            loading,
        }
    }

    const namespaces = data.viewer.affiliatedNamespaces.nodes
    const initialNamespace = initialNamespaceID
        ? namespaces.find(ns => ns.id === initialNamespaceID)
        : namespaces.find(ns => ns.__typename === 'User') ?? namespaces.at(0)

    return {
        namespaces,
        loading: false,
        initialNamespace,
    }
}

export const viewerAffiliatedNamespacesQuery = gql`
    query ViewerAffiliatedNamespaces {
        viewer {
            affiliatedNamespaces {
                nodes {
                    __typename
                    id
                    namespaceName
                }
            }
        }
    }
`
