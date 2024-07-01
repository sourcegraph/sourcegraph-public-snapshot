/**
 * Common props for components underneath a namespace (e.g., a user or organization).
 */
export interface NamespaceProps {
    /**
     * The namespace.
     */
    namespace: { __typename: 'User' | 'Org'; id: string; url: string }
}
