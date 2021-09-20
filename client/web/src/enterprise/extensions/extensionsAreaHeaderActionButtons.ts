import AddIcon from 'mdi-react/AddIcon'

import {
    ExtensionsAreaHeaderActionButton,
    extensionsAreaHeaderActionButtons,
} from '../../extensions/registry/ExtensionsAreaHeader'

export const enterpriseExtensionsAreaHeaderActionButtons: readonly ExtensionsAreaHeaderActionButton[] = [
    ...extensionsAreaHeaderActionButtons,
    {
        label: 'Create extension',
        to: () => '/extensions/registry/new',
        icon: AddIcon,
        condition: context => context.isPrimaryHeader,
    },
]
