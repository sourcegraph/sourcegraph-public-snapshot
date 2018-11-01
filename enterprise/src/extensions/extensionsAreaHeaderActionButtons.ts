import AddIcon from 'mdi-react/AddIcon'
import { ExtensionsAreaHeaderActionButton } from '../../../packages/webapp/src/extensions/ExtensionsAreaHeader'
import { extensionsAreaHeaderActionButtons } from '../../../packages/webapp/src/extensions/extensionsAreaHeaderActionButtons'

export const enterpriseExtensionsAreaHeaderActionButtons: ReadonlyArray<ExtensionsAreaHeaderActionButton> = [
    ...extensionsAreaHeaderActionButtons,
    {
        label: 'Publish new extension',
        to: () => '/extensions/registry/new',
        icon: AddIcon,
        condition: context => context.isPrimaryHeader,
    },
]
