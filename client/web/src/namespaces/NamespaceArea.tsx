import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { AuthenticatedUser } from '../auth'
import { BatchChangesProps } from '../batches'
import { UserAreaUserFields, OrgAreaOrganizationFields } from '../graphql-operations'
import { NavItemWithIconDescriptor, RouteDescriptor } from '../util/contributions'

/**
 * Properties passed to all page components in the namespace area.
 */
export interface NamespaceAreaContext extends ExtensionsControllerProps, ThemeProps {
    namespace: Pick<UserAreaUserFields | OrgAreaOrganizationFields, '__typename' | 'id' | 'url'>

    authenticatedUser: AuthenticatedUser | null
}

export interface NamespaceAreaRoute extends RouteDescriptor<NamespaceAreaContext> {}

interface NavItemDescriptorContext extends BatchChangesProps {
    isSourcegraphDotCom: boolean
}

export interface NamespaceAreaNavItem extends NavItemWithIconDescriptor<NavItemDescriptorContext> {}
