import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { RouteDescriptor } from '../util/contributions'
import { PatternTypeProps } from '../search'
import { ThemeProps } from '../../../shared/src/theme'

/**
 * Properties passed to all page components in the namespace area.
 */
export interface NamespaceAreaContext
    extends ExtensionsControllerProps,
        ThemeProps,
        Omit<PatternTypeProps, 'setPatternType'> {
    namespace: Pick<GQL.Namespace, '__typename' | 'id' | 'url'>

    authenticatedUser: GQL.IUser | null
}

export interface NamespaceAreaRoute extends RouteDescriptor<NamespaceAreaContext> {}
