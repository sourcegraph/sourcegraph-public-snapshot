export function getCommandPaletteMount(): HTMLElement {
    const headerElem = document.querySelector('div.HeaderMenu>div:last-child')
    if (!headerElem) {
        throw new Error('Unable to find command palette mount')
    }

    const commandListClass = 'command-palette-button'

    const createCommandList = (): HTMLElement => {
        const commandListElem = document.createElement('div')
        commandListElem.className = commandListClass
        headerElem.insertAdjacentElement('afterbegin', commandListElem)

        return commandListElem
    }

    return document.querySelector<HTMLElement>('.' + commandListClass) || createCommandList()
}

export function getGlobalDebugMount(): HTMLElement {
    const globalDebugClass = 'global-debug'

    const createGlobalDebugMount = (): HTMLElement => {
        const globalDebugElem = document.createElement('div')
        globalDebugElem.className = globalDebugClass
        document.body.appendChild(globalDebugElem)

        return globalDebugElem
    }

    return document.querySelector<HTMLElement>('.' + globalDebugClass) || createGlobalDebugMount()
}
