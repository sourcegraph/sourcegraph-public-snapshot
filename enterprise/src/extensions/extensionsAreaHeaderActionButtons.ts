import { ExtensionsAreaHeaderActionButton } from '@sourcegraph/webapp/dist/extensions/ExtensionsAreaHeader'
import { extensionsAreaHeaderActionButtons } from '@sourcegraph/webapp/dist/extensions/extensionsAreaHeaderActionButtons'
import AddIcon from 'mdi-react/AddIcon'

export const enterpriseExtensionsAreaHeaderActionButtons: ReadonlyArray<ExtensionsAreaHeaderActionButton> = [
    ...extensionsAreaHeaderActionButtons,
    {
        label: 'Publish new extension',
        to: () => '/extensions/registry/new',
        icon: AddIcon,
        condition: context => context.isPrimaryHeader,
    },
]
