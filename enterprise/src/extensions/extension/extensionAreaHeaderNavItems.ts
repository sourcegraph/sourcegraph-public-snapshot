import LockIcon from 'mdi-react/LockIcon'
import { ExtensionAreaHeaderNavItem } from '../../../../packages/webapp/src/extensions/extension/ExtensionAreaHeader'
import { extensionAreaHeaderNavItems } from '../../../../packages/webapp/src/extensions/extension/extensionAreaHeaderNavItems'

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
