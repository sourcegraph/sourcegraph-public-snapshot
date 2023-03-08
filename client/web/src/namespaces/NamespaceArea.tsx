import { AuthenticatedUser } from '../auth'
import { BatchChangesProps } from '../batches'
import { UserAreaUserFields, OrgAreaOrganizationFields } from '../graphql-operations'
import { NavItemWithIconDescriptor, RouteV6Descriptor } from '../util/contributions'

/**
 * Properties passed to all page components in the namespace area.
 */
export interface NamespaceAreaContext {
    namespace: Pick<UserAreaUserFields | OrgAreaOrganizationFields, '__typename' | 'id' | 'url'>

    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
}

export interface NamespaceAreaRoute extends RouteV6Descriptor<NamespaceAreaContext> {}

interface NavItemDescriptorContext extends BatchChangesProps {
    isSourcegraphDotCom: boolean
    isSourcegraphApp: boolean
}

export interface NamespaceAreaNavItem extends NavItemWithIconDescriptor<NavItemDescriptorContext> {}
