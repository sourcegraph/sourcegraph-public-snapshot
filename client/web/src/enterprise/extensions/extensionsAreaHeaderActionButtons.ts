import AddIcon from 'mdi-react/AddIcon'

import type { ExtensionsAreaHeaderActionButton } from '../../extensions/ExtensionsAreaHeader'
import { extensionsAreaHeaderActionButtons } from '../../extensions/extensionsAreaHeaderActionButtons'

export const enterpriseExtensionsAreaHeaderActionButtons: readonly ExtensionsAreaHeaderActionButton[] = [
    ...extensionsAreaHeaderActionButtons,
    {
        label: 'Create extension',
        to: () => '/extensions/registry/new',
        icon: AddIcon,
        condition: context => context.isPrimaryHeader,
    },
]
