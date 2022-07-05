import { mdiPlus } from '@mdi/js'

import { ExtensionsAreaHeaderActionButton } from '../../extensions/ExtensionsAreaHeader'
import { extensionsAreaHeaderActionButtons } from '../../extensions/extensionsAreaHeaderActionButtons'

export const enterpriseExtensionsAreaHeaderActionButtons: readonly ExtensionsAreaHeaderActionButton[] = [
    ...extensionsAreaHeaderActionButtons,
    {
        label: 'Create extension',
        to: () => '/extensions/registry/new',
        icon: mdiPlus,
        condition: context => context.isPrimaryHeader,
    },
]
