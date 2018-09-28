import { CommandListPopoverButton } from '@sourcegraph/extensions-client-common/lib/app/CommandList'
import { Controller as ClientController } from '@sourcegraph/extensions-client-common/lib/client/controller'
import { Controller } from '@sourcegraph/extensions-client-common/lib/controller'
import { ConfigurationSubject, Settings } from '@sourcegraph/extensions-client-common/lib/settings'
import { ContributableMenu } from 'sourcegraph/module/protocol'

import * as React from 'react'
import { render } from 'react-dom'

export function getCommandPaletteMount(): HTMLElement {
    const headerElem = document.querySelector('div.HeaderMenu>div:last-child')
    if (!headerElem) {
        throw new Error('Unable to find command pallete mount')
    }

    const commandListClass = 'command-palette-button'

    const createCommandList = (): HTMLElement => {
        const commandListElem = document.createElement('div')
        commandListElem.className = commandListClass
        headerElem!.appendChild(commandListElem)

        return commandListElem
    }

    return document.querySelector<HTMLElement>('.' + commandListClass) || createCommandList()
}

// TODO: remove with old inject
export function injectExtensionsGlobalComponents({
    extensionsController,
    extensionsContextController,
}: {
    extensionsController: ClientController<ConfigurationSubject, Settings>
    extensionsContextController: Controller<ConfigurationSubject, Settings>
}): void {
    render(
        <CommandListPopoverButton
            extensionsController={extensionsController}
            menu={ContributableMenu.CommandPalette}
            extensions={extensionsContextController}
        />,
        getCommandPaletteMount()
    )
}
