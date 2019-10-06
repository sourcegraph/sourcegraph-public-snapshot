import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { ThemeProps } from '../theme'
import { RouteDescriptor } from '../util/contributions'
import { PlatformContextProps } from '../../../shared/src/platform/context'

/**
 * Properties passed to all page components in the namespace area.
 */
export interface NamespaceAreaContext extends ExtensionsControllerProps, PlatformContextProps, ThemeProps {
    namespace: Pick<GQL.Namespace, '__typename' | 'id' | 'namespaceName' | 'url'>

    authenticatedUser: GQL.IUser | null
}

export interface NamespaceAreaRoute extends RouteDescriptor<NamespaceAreaContext> {}
