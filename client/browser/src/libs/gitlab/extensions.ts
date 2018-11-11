export function getCommandPaletteMount(): HTMLElement {
    const headerElem = document.querySelector('.navbar-collapse')
    if (!headerElem) {
        throw new Error('Unable to find command pallete mount')
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
