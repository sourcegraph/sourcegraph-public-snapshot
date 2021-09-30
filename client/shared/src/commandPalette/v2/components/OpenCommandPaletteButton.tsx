import React from 'react'
import ConsoleIcon from 'mdi-react/ConsoleIcon'
import { useCommandPaletteStore } from '../store'
import classNames from 'classnames'
import { Key } from 'ts-key-enum'
import styles from './OpenCommandPaletteButton.module.scss'

export interface OpenCommandPaletteButtonProps {
    buttonClassName?: string
    buttonElement?: 'span' | 'a'
}

export const OpenCommandPaletteButton: React.FC<OpenCommandPaletteButtonProps> = ({
    buttonClassName,
    buttonElement: ButtonElement = 'span',
}) => {
    const toggleCommandPaletteIsOpen = useCommandPaletteStore(state => state.toggleIsOpen)

    const onKeyDown: React.KeyboardEventHandler<HTMLSpanElement> = event => {
        if (event.key === Key.Enter) {
            toggleCommandPaletteIsOpen()
        }
    }

    return (
        <ButtonElement
            role="button"
            className={classNames(buttonClassName, styles.button)}
            onClick={toggleCommandPaletteIsOpen as () => void}
            tabIndex={0}
            onKeyDown={onKeyDown}
            data-tooltip="Open command palette"
        >
            <ConsoleIcon className="icon-inline-md" />
        </ButtonElement>
    )
}
