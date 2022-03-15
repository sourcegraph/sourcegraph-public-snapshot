import React, { useRef } from 'react'

import classNames from 'classnames'

import { BlockProps } from '..'

import { NotebookBlockMenu, NotebookBlockMenuProps } from './menu/NotebookBlockMenu'
import { useBlockSelection } from './useBlockSelection'
import { useBlockShortcuts } from './useBlockShortcuts'

import blockStyles from './NotebookBlock.module.scss'

interface NotebookBlockProps extends Omit<BlockProps, 'input' | 'output'>, NotebookBlockMenuProps {
    className?: string
    isInputFocused: boolean
    'aria-label': string
    onDoubleClick?: () => void
    onEnterBlock: () => void
}

export const NotebookBlock: React.FunctionComponent<NotebookBlockProps> = ({
    children,
    id,
    className,
    isSelected,
    isInputFocused,
    isOtherBlockSelected,
    mainAction,
    actions,
    'aria-label': ariaLabel,
    onEnterBlock,
    onDoubleClick,
    ...props
}) => {
    const blockElement = useRef(null)

    const { onSelect } = useBlockSelection({
        id,
        blockElement: blockElement.current,
        isSelected,
        isInputFocused,
        ...props,
    })

    const { onKeyDown } = useBlockShortcuts({
        id,
        onEnterBlock,
        ...props,
    })

    return (
        <div className={classNames('block-wrapper', blockStyles.blockWrapper)} data-block-id={id}>
            {/* Notebook blocks are a form of specialized UI for which there are no good accesibility settings (role, aria-*)
            or semantic elements that would accurately describe its functionality. To provide the necessary functionality we have
            to rely on plain div elements and custom click/focus/keyDown handlers. We still preserve the ability to navigate through blocks
            with the keyboard using the up and down arrows, and TAB. */}
            {/* eslint-disable-next-line jsx-a11y/no-static-element-interactions */}
            <div
                className={classNames(
                    blockStyles.block,
                    className,
                    isSelected && !isInputFocused && blockStyles.selected,
                    isSelected && isInputFocused && blockStyles.selectedNotFocused
                )}
                onClick={onSelect}
                onDoubleClick={onDoubleClick}
                onKeyDown={onKeyDown}
                onFocus={onSelect}
                // A tabIndex is necessary to make the block focusable.
                // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
                tabIndex={0}
                aria-label={ariaLabel}
                ref={blockElement}
            >
                {children}
            </div>
            {(isSelected || !isOtherBlockSelected) && (
                <NotebookBlockMenu id={id} mainAction={mainAction} actions={actions} />
            )}
        </div>
    )
}
