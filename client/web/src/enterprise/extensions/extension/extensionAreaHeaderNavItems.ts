import LockIcon from 'mdi-react/LockIcon'

import { ExtensionAreaHeaderNavItem } from '../../../extensions/extension/ExtensionAreaHeader'

export const enterpriseExtensionAreaHeaderNavItems: readonly ExtensionAreaHeaderNavItem[] = [
    {
        to: '',
        exact: true,
        label: 'Extension',
    },
    {
        to: '/-/contributions',
        exact: true,
        label: 'Contributions',
    },
    {
        condition: context =>
            !!context.extension.registryExtension && context.extension.registryExtension.viewerCanAdminister,
        to: '/-/manage',
        icon: LockIcon,
        label: 'Manage',
    },
]
