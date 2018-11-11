import AddIcon from 'mdi-react/AddIcon'
import { ExtensionsAreaHeaderActionButton } from '../../../web/src/extensions/ExtensionsAreaHeader'
import { extensionsAreaHeaderActionButtons } from '../../../web/src/extensions/extensionsAreaHeaderActionButtons'

export const enterpriseExtensionsAreaHeaderActionButtons: ReadonlyArray<ExtensionsAreaHeaderActionButton> = [
    ...extensionsAreaHeaderActionButtons,
    {
        label: 'Publish new extension',
        to: () => '/extensions/registry/new',
        icon: AddIcon,
        condition: context => context.isPrimaryHeader,
    },
]
