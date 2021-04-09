import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { NavItemWithIconDescriptor, RouteDescriptor } from '../util/contributions'
import { PatternTypeProps } from '../search'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { AuthenticatedUser } from '../auth'

/**
 * Properties passed to all page components in the namespace area.
 */
export interface NamespaceAreaContext
    extends ExtensionsControllerProps,
        ThemeProps,
        Omit<PatternTypeProps, 'setPatternType'> {
    namespace: Pick<GQL.Namespace, '__typename' | 'id' | 'url'>

    authenticatedUser: AuthenticatedUser | null
}

export interface NamespaceAreaRoute extends RouteDescriptor<NamespaceAreaContext> {}

export interface NamespaceAreaNavItem
    extends NavItemWithIconDescriptor<{
        isSourcegraphDotCom: boolean
    }> {}
