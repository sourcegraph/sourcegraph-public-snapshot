import { RouteComponentProps } from 'react-router-dom'

import { NavItemWithIconDescriptor } from '../../util/contributions'

import { ExtensionAreaRouteContext } from './ExtensionArea'

export interface ExtensionAreaHeaderProps extends ExtensionAreaRouteContext, RouteComponentProps<{}> {
    navItems: readonly ExtensionAreaHeaderNavItem[]
    className?: string
}

export type ExtensionAreaHeaderContext = Pick<ExtensionAreaHeaderProps, 'extension'>

export interface ExtensionAreaHeaderNavItem extends NavItemWithIconDescriptor<ExtensionAreaHeaderContext> {}

export const extensionAreaHeaderNavItems: ExtensionAreaHeaderNavItem[] = []
