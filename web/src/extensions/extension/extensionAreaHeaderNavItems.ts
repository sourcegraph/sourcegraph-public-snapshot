import { LockIcon } from 'mdi-react'
import { ExtensionAreaHeaderNavItem } from './ExtensionAreaHeader'

export const extensionAreaHeaderNavItems: ReadonlyArray<ExtensionAreaHeaderNavItem> = [
    {
        to: '',
        exact: true,
        label: 'Extension',
    },
    {
        condition: context =>
            !!context.extension.registryExtension && context.extension.registryExtension.viewerCanAdminister,
        to: '/-/manage',
        icon: LockIcon,
        label: 'Manage',
    },
]
