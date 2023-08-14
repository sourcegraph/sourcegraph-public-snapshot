import React, { useCallback, useEffect, useMemo } from 'react'

import classNames from 'classnames'

import { isMacPlatform as isMacPlatformFunc } from '@sourcegraph/common'
import { isInputElement } from '@sourcegraph/shared/src/util/dom'

import type { BlockProps } from '..'
import { isModifierKeyPressed } from '../notebook/useNotebookEventHandlers'

import { NotebookBlockMenu, type NotebookBlockMenuProps } from './menu/NotebookBlockMenu'
import { useIsBlockInputFocused } from './useIsBlockInputFocused'

import blockStyles from './NotebookBlock.module.scss'

interface NotebookBlockProps extends Pick<BlockProps, 'isSelected' | 'showMenu'>, NotebookBlockMenuProps {
    className?: string
    'aria-label': string
    onDoubleClick?: () => void
    isReadOnly: boolean
    isInputVisible?: boolean
    setIsInputVisible?: (value: boolean) => void
    focusInput?: () => void
}

export const NotebookBlock: React.FunctionComponent<React.PropsWithChildren<NotebookBlockProps>> = ({
    children,
    id,
    className,
    isSelected,
    showMenu,
    mainAction,
    actions,
    'aria-label': ariaLabel,
    onDoubleClick,
    isReadOnly,
    isInputVisible,
    setIsInputVisible,
    focusInput,
}) => {
    const isInputFocused = useIsBlockInputFocused(id)
    const isMacPlatform = useMemo(() => isMacPlatformFunc(), [])

    const onEnterBlock = useCallback(() => {
        if (isInputVisible) {
            focusInput?.()
        } else if (!isReadOnly) {
            setIsInputVisible?.(true)
        }
    }, [isInputVisible, isReadOnly, focusInput, setIsInputVisible])

    useEffect(() => {
        const handleKeyDown = (event: KeyboardEvent): void => {
            const target = event.target as HTMLElement
            if (isInputElement(target)) {
                return
            }

            if (isSelected && event.key === 'Enter') {
                if (isModifierKeyPressed(event.metaKey, event.ctrlKey, isMacPlatform)) {
                    setIsInputVisible?.(false)
                } else {
                    // This prevents CodeMirror from appending a new line when
                    // the input was focused by pressing the Enter key
                    event.preventDefault()
                    onEnterBlock()
                }
            }
        }

        document.addEventListener('keydown', handleKeyDown)
        return () => {
            document.removeEventListener('keydown', handleKeyDown)
        }
    }, [isMacPlatform, isSelected, onEnterBlock, setIsInputVisible])

    return (
        <div className={classNames('block-wrapper', blockStyles.blockWrapper)} data-block-id={id}>
            {/* Notebook blocks are a form of specialized UI. Since they are items in a list that use some common UI,
            we can use the `article` semantic element. To provide the necessary functionality we have
            to provide custom click/focus/keyDown handlers for arrows/enter, as well as setting a tabindex. */}
            <article
                className={classNames(
                    'block',
                    blockStyles.block,
                    className,
                    isSelected && !isInputFocused && blockStyles.selected,
                    isSelected && isInputFocused && blockStyles.selectedNotFocused
                )}
                onDoubleClick={onDoubleClick}
                // A tabIndex is necessary to make the block focusable.
                // ARC Toolkit will complain about this, but setting a
                // role (implicitly via use of `article`) and aria-label
                // is valid by WCAG 2.1 Success Criterion 4.1.2.
                // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
                tabIndex={0}
                aria-label={ariaLabel}
            >
                {children}
            </article>
            {showMenu && <NotebookBlockMenu id={id} mainAction={mainAction} actions={actions} />}
        </div>
    )
}
