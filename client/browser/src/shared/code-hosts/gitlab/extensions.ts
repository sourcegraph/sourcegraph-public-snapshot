import { querySelectorOrSelf } from '../../util/dom'
import { MountGetter } from '../shared/codeHost'

import styles from './codeHost.module.scss'

export const getCommandPaletteMount: MountGetter = (container: HTMLElement): HTMLElement | null => {
    const headerElement = querySelectorOrSelf(container, '.navbar-collapse')
    if (!headerElement) {
        return null
    }
    const commandListClass = 'command-palette-button'
    const createCommandList = (): HTMLElement => {
        const mount = document.createElement('div')
        mount.classList.add(commandListClass, styles.commandPaletteButton)
        headerElement.prepend(mount)
        return mount
    }
    return headerElement.querySelector<HTMLElement>('.' + commandListClass) || createCommandList()
}
