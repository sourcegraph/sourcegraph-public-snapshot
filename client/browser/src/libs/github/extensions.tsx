import { querySelectorOrSelf } from '../../shared/util/dom'
import { MountGetter } from '../code_intelligence'

export const getCommandPaletteMount: MountGetter = (container: HTMLElement): HTMLElement | null => {
    const headerElement = querySelectorOrSelf(container, 'div.HeaderMenu > div:last-child')
    if (!headerElement) {
        return null
    }
    const className = 'command-palette-button'
    const createCommandPaletteMount = (): HTMLElement => {
        const mount = document.createElement('div')
        mount.className = className
        headerElement.insertAdjacentElement('afterbegin', mount)
        return mount
    }
    return headerElement.querySelector<HTMLElement>('.' + className) || createCommandPaletteMount()
}

export const getGlobalDebugMount: MountGetter = (container: HTMLElement): HTMLElement | null => {
    const globalDebugClass = 'global-debug'
    const parentElement = querySelectorOrSelf(container, 'body')
    if (!parentElement) {
        return null
    }
    const createGlobalDebugMount = (): HTMLElement => {
        const mount = document.createElement('div')
        mount.className = globalDebugClass
        parentElement.appendChild(mount)
        return mount
    }
    return container.querySelector<HTMLElement>('.' + globalDebugClass) || createGlobalDebugMount()
}
