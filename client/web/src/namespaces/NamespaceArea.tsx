import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { NavItemWithIconDescriptor, RouteDescriptor } from '../util/contributions'
import { PatternTypeProps } from '../search'
import { ThemeProps } from '../../../shared/src/theme'
import { AuthenticatedUser } from '../auth'
import { GraphSelectionProps } from '../enterprise/graphs/selector/graphSelectionProps'

/**
 * Properties passed to all page components in the namespace area.
 */
export interface NamespaceAreaContext
    extends ExtensionsControllerProps,
        ThemeProps,
        Pick<GraphSelectionProps, 'reloadGraphs'>,
        Omit<PatternTypeProps, 'setPatternType'> {
    namespace: Pick<GQL.Namespace, '__typename' | 'id' | 'url'>

    authenticatedUser: AuthenticatedUser | null
}

export interface NamespaceAreaRoute extends RouteDescriptor<NamespaceAreaContext> {
    hideNamespaceAreaSidebar?: boolean
}

export interface NamespaceAreaNavItem
    extends NavItemWithIconDescriptor<{
        isSourcegraphDotCom: boolean
    }> {}
