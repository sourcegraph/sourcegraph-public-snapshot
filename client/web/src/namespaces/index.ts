import type { OrgAreaOrganizationFields, UserAreaUserFields } from '../graphql-operations'

/**
 * Common props for components underneath a namespace (e.g., a user or organization).
 */
export interface NamespaceProps {
    /**
     * The namespace.
     */
    namespace: Pick<UserAreaUserFields | OrgAreaOrganizationFields, '__typename' | 'id' | 'url'>
}
