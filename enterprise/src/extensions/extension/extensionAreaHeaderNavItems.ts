import { ExtensionAreaHeaderNavItem } from '@sourcegraph/webapp/dist/extensions/extension/ExtensionAreaHeader'
import { extensionAreaHeaderNavItems } from '@sourcegraph/webapp/dist/extensions/extension/extensionAreaHeaderNavItems'
import LockIcon from 'mdi-react/LockIcon'

export const enterpriseExtensionAreaHeaderNavItems: ReadonlyArray<ExtensionAreaHeaderNavItem> = [
    ...extensionAreaHeaderNavItems,
    {
        condition: context =>
            !!context.extension.registryExtension && context.extension.registryExtension.viewerCanAdminister,
        to: '/-/manage',
        icon: LockIcon,
        label: 'Manage',
    },
]
