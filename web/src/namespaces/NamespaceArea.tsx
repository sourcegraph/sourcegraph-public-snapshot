import { ExtensionsControllerNotificationProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { ThemeProps } from '../theme'
import { RouteDescriptor } from '../util/contributions'

/**
 * Properties passed to all page components in the namespace area.
 */
export interface NamespaceAreaContext extends ExtensionsControllerNotificationProps, ThemeProps {
    namespace: Pick<GQL.Namespace, '__typename' | 'id' | 'url'>
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

export interface NamespaceAreaRoute extends RouteDescriptor<NamespaceAreaContext> {}
