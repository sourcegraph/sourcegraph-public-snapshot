import { querySelectorOrSelf } from '../../util/dom'
import { MountGetter } from '../shared/codeHost'

export const getCommandPaletteMount: MountGetter = (container: HTMLElement): HTMLElement | null => {
    const headerElem = querySelectorOrSelf(container, '.navbar-collapse')
    if (!headerElem) {
        return null
    }
    const commandListClass = 'command-palette-button'
    const createCommandList = (): HTMLElement => {
        const mount = document.createElement('div')
        mount.className = commandListClass
        headerElem.prepend(mount)
        return mount
    }
    return headerElem.querySelector<HTMLElement>('.' + commandListClass) || createCommandList()
}
