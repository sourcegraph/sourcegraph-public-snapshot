import classNames from 'classnames'
import { noop } from 'lodash'
import PencilIcon from 'mdi-react/PencilIcon'
import PlayCircleOutlineIcon from 'mdi-react/PlayCircleOutlineIcon'
import * as Monaco from 'monaco-editor'
import React, { useState, useCallback, useRef, useEffect, useMemo } from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { MonacoEditor } from '@sourcegraph/shared/src/components/MonacoEditor'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Icon } from '@sourcegraph/wildcard'

import { BlockProps, MarkdownBlock } from '../..'
import { BlockMenuAction, NotebookBlockMenu } from '../menu/NotebookBlockMenu'
import { useCommonBlockMenuActions } from '../menu/useCommonBlockMenuActions'
import blockStyles from '../NotebookBlock.module.scss'
import { useBlockSelection } from '../useBlockSelection'
import { useBlockShortcuts } from '../useBlockShortcuts'
import { MONACO_BLOCK_INPUT_OPTIONS, useMonacoBlockInput } from '../useMonacoBlockInput'

import styles from './NotebookMarkdownBlock.module.scss'

interface NotebookMarkdownBlockProps extends BlockProps, MarkdownBlock, ThemeProps {
    isMacPlatform: boolean
}

export const NotebookMarkdownBlock: React.FunctionComponent<NotebookMarkdownBlockProps> = ({
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
    const [isEditing, setIsEditing] = useState(!isReadOnly && input.length === 0)
    const [editor, setEditor] = useState<Monaco.editor.IStandaloneCodeEditor>()
    const blockElement = useRef<HTMLDivElement>(null)

    const runBlock = useCallback(
        (id: string) => {
            onRunBlock(id)
            setIsEditing(false)
        },
        [onRunBlock, setIsEditing]
    )

    const { isInputFocused } = useMonacoBlockInput({
        editor,
        id,
        ...props,
        onSelectBlock,
        onRunBlock: runBlock,
    })

    const onDoubleClick = useCallback(() => {
        if (isReadOnly) {
            return
        }
        if (!isEditing) {
            setIsEditing(true)
            onSelectBlock(id)
        }
    }, [id, isReadOnly, isEditing, setIsEditing, onSelectBlock])

    // setTimeout turns on editing mode in a separate run-loop which prevents adding a newline at the start of the input
    const onEnterBlock = useCallback(() => {
        if (isReadOnly) {
            return
        }
        setTimeout(() => setIsEditing(true), 0)
    }, [isReadOnly, setIsEditing])

    const { onSelect } = useBlockSelection({
        id,
        blockElement: blockElement.current,
        isSelected,
        isInputFocused,
        onSelectBlock,
        ...props,
    })

    const { onKeyDown } = useBlockShortcuts({ id, isMacPlatform, onEnterBlock, ...props, onRunBlock: runBlock })

    useEffect(() => {
        if (isEditing) {
            editor?.focus()
        }
    }, [isEditing, editor])

    const modifierKeyLabel = isMacPlatform ? '⌘' : 'Ctrl'
    const commonMenuActions = useCommonBlockMenuActions({
        modifierKeyLabel,
        isInputFocused,
        isReadOnly,
        isMacPlatform,
        ...props,
    })
    const menuActions = useMemo(() => {
        const action: BlockMenuAction[] = [
            isEditing
                ? {
                      type: 'button',
                      label: 'Render',
                      icon: <Icon as={PlayCircleOutlineIcon} />,
                      onClick: runBlock,
                      keyboardShortcutLabel: `${modifierKeyLabel} + ↵`,
                  }
                : {
                      type: 'button',
                      label: 'Edit',
                      icon: <Icon as={PencilIcon} />,
                      onClick: onEnterBlock,
                      keyboardShortcutLabel: '↵',
                  },
        ]
        return action.concat(commonMenuActions)
    }, [isEditing, modifierKeyLabel, runBlock, onEnterBlock, commonMenuActions])

    const blockMenu = isSelected && !isReadOnly && <NotebookBlockMenu id={id} actions={menuActions} />

    if (!isEditing) {
        return (
            <div className={classNames('block-wrapper', blockStyles.blockWrapper)} data-block-id={id}>
                {/* Notebook blocks are a form of specialized UI for which there are no good accesibility settings (role, aria-*)
                    or semantic elements that would accurately describe its functionality. To provide the necessary functionality we have
                    to rely on plain div elements and custom click/focus/keyDown handlers. We still preserve the ability to navigate through blocks
                    with the keyboard using the up and down arrows, and TAB. */}
                {/* eslint-disable-next-line jsx-a11y/no-static-element-interactions */}
                <div
                    className={classNames(blockStyles.block, isSelected && blockStyles.selected, styles.outputWrapper)}
                    onClick={onSelect}
                    onFocus={onSelect}
                    onDoubleClick={onDoubleClick}
                    onKeyDown={onKeyDown}
                    // A tabIndex is necessary to make the block focusable.
                    // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
                    tabIndex={0}
                    aria-label="Notebook markdown block"
                    ref={blockElement}
                >
                    <div className={styles.output} data-testid="output">
                        <Markdown className={styles.markdown} dangerousInnerHTML={output ?? ''} />
                    </div>
                </div>
                {blockMenu}
            </div>
        )
    }

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
                aria-label="Notebook markdown block"
                ref={blockElement}
            >
                <div className={blockStyles.monacoWrapper}>
                    <MonacoEditor
                        language="markdown"
                        value={input}
                        height="auto"
                        isLightTheme={isLightTheme}
                        editorWillMount={noop}
                        onEditorCreated={setEditor}
                        options={MONACO_BLOCK_INPUT_OPTIONS}
                        border={false}
                    />
                </div>
            </div>
            {blockMenu}
        </div>
    )
}
