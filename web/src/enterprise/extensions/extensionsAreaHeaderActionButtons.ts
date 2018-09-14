import AddIcon from 'mdi-react/AddIcon'
import { extensionsAreaHeaderActionButtons } from '../../extensions/extensionAreaHeaderActionButtons'
import { ExtensionsAreaHeaderActionButton } from '../../extensions/ExtensionsAreaHeader'

export const enterpriseExtensionsAreaHeaderActionButtons: ReadonlyArray<ExtensionsAreaHeaderActionButton> = [
    ...extensionsAreaHeaderActionButtons,
    {
        label: 'Publish new extension',
        to: () => '/extensions/registry/new',
        icon: AddIcon,
        condition: context => context.isPrimaryHeader,
    },
]
