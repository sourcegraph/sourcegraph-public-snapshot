import classNames from 'classnames'
import React, { useRef } from 'react'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BlockProps, ComputeBlock } from '../..'
import { NotebookBlockMenu } from '../menu/NotebookBlockMenu'
import { useCommonBlockMenuActions } from '../menu/useCommonBlockMenuActions'
import blockStyles from '../NotebookBlock.module.scss'
import { useBlockSelection } from '../useBlockSelection'
import { useBlockShortcuts } from '../useBlockShortcuts'

import styles from './NotebookComputeBlock.module.scss'

interface ComputeBlockProps extends BlockProps, ComputeBlock, ThemeProps {
    isMacPlatform: boolean
}

export const NotebookComputeBlock: React.FunctionComponent<ComputeBlockProps> = ({
    id,
    input,
    output,
    isSelected,
    isLightTheme,
    isMacPlatform,
    isReadOnly,
    onRunBlock,
    onSelectBlock,
    ...props
}) => {
    const isInputFocused = false
    const blockElement = useRef<HTMLDivElement>(null)

    const { onSelect } = useBlockSelection({
        id,
        blockElement: blockElement.current,
        isSelected,
        isInputFocused,
        onSelectBlock,
        ...props,
    })

    const { onKeyDown } = useBlockShortcuts({
        id,
        isMacPlatform,
        onEnterBlock: () => {},
        ...props,
        onRunBlock: () => {},
    })

    const modifierKeyLabel = isMacPlatform ? '⌘' : 'Ctrl'
    const commonMenuActions = useCommonBlockMenuActions({
        modifierKeyLabel,
        isInputFocused,
        isReadOnly,
        isMacPlatform,
        ...props,
    })

    const blockMenu = isSelected && !isReadOnly && <NotebookBlockMenu id={id} actions={commonMenuActions} />

    return (
        <div className={classNames('block-wrapper', blockStyles.blockWrapper)} data-block-id={id}>
            {/* See the explanation for the disable above. */}
            {/* eslint-disable-next-line jsx-a11y/no-static-element-interactions */}
            <div
                className={classNames(
                    blockStyles.block,
                    styles.input,
                    (isInputFocused || isSelected) && blockStyles.selected
                )}
                onClick={onSelect}
                onFocus={onSelect}
                onKeyDown={onKeyDown}
                // A tabIndex is necessary to make the block focusable.
                // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
                tabIndex={0}
                aria-label="Notebook compute block"
                ref={blockElement}
            >
                <div className="elm" />
            </div>
            {blockMenu}
        </div>
    )
}
