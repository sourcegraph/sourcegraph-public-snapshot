import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'

import type { AuthenticatedUser } from '../auth'
import type { BatchChangesProps } from '../batches'
import type { NavItemWithIconDescriptor, RouteV6Descriptor } from '../util/contributions'

import type { NamespaceProps } from '.'

/**
 * Properties passed to all page components in the namespace area.
 */
export interface NamespaceAreaContext extends PlatformContextProps, NamespaceProps {
    authenticatedUser: AuthenticatedUser | null
}

export interface NamespaceAreaRoute extends RouteV6Descriptor<NamespaceAreaContext> {}

interface NavItemDescriptorContext extends BatchChangesProps {}

export interface NamespaceAreaNavItem extends NavItemWithIconDescriptor<NavItemDescriptorContext> {}
