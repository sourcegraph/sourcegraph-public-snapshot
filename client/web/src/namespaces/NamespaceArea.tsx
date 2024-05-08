import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'

import type { AuthenticatedUser } from '../auth'
import type { BatchChangesProps } from '../batches'
import type { UserAreaUserFields, OrgAreaOrganizationFields } from '../graphql-operations'
import type { NavItemWithIconDescriptor, RouteV6Descriptor } from '../util/contributions'

/**
 * Properties passed to all page components in the namespace area.
 */
export interface NamespaceAreaContext extends PlatformContextProps {
    namespace: Pick<UserAreaUserFields | OrgAreaOrganizationFields, '__typename' | 'id' | 'url'>

    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
    license: {
        isCodeSearchEnabled: boolean
        isCodyEnabled: boolean
    }
}

export interface NamespaceAreaRoute extends RouteV6Descriptor<NamespaceAreaContext> {}

interface NavItemDescriptorContext extends BatchChangesProps {
    isSourcegraphDotCom: boolean
}

export interface NamespaceAreaNavItem extends NavItemWithIconDescriptor<NavItemDescriptorContext> {}
