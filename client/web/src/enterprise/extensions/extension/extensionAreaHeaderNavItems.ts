import { mdiLock } from '@mdi/js'

import { ExtensionAreaHeaderNavItem } from '../../../extensions/extension/ExtensionAreaHeader'
import { extensionAreaHeaderNavItems } from '../../../extensions/extension/extensionAreaHeaderNavItems'

export const enterpriseExtensionAreaHeaderNavItems: readonly ExtensionAreaHeaderNavItem[] = [
    ...extensionAreaHeaderNavItems,
    {
        condition: context =>
            !!context.extension.registryExtension && context.extension.registryExtension.viewerCanAdminister,
        to: '/-/manage',
        icon: mdiLock,
        label: 'Manage',
    },
]
