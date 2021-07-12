import classNames from 'classnames'
import { noop } from 'lodash'
import PencilIcon from 'mdi-react/PencilIcon'
import PlayCircleOutlineIcon from 'mdi-react/PlayCircleOutlineIcon'
import * as Monaco from 'monaco-editor'
import React, { useState, useCallback, useRef, useEffect, useMemo } from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { MonacoEditor } from '@sourcegraph/web/src/components/MonacoEditor'

import blockStyles from './SearchNotebookBlock.module.scss'
import { SearchNotebookBlockMenu } from './SearchNotebookBlockMenu'
import styles from './SearchNotebookMarkdownBlock.module.scss'
import { useBlockSelection } from './useBlockSelection'
import { useBlockShortcuts } from './useBlockShortcuts'
import { useCommonBlockMenuActions } from './useCommonBlockMenuActions'
import { MONACO_BLOCK_INPUT_OPTIONS, useMonacoBlockInput } from './useMonacoBlockInput'

import { BlockProps, MarkdownBlock } from '.'

interface SearchNotebookMarkdownBlockProps extends BlockProps, Omit<MarkdownBlock, 'type'>, ThemeProps {
    isMacPlatform: boolean
}

export const SearchNotebookMarkdownBlock: React.FunctionComponent<SearchNotebookMarkdownBlockProps> = ({
    id,
    input,
    output,
    isSelected,
    isLightTheme,
    isMacPlatform,
    onRunBlock,
    onBlockInputChange,
    onSelectBlock,
    onMoveBlockSelection,
    onDeleteBlock,
    onMoveBlock,
    onDuplicateBlock,
}) => {
    const [isEditing, setIsEditing] = useState(false)
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
        onRunBlock: runBlock,
        onBlockInputChange,
        onSelectBlock,
        onMoveBlockSelection,
    })

    const onDoubleClick = useCallback(() => {
        if (!isEditing) {
            setIsEditing(true)
            onSelectBlock(id)
        }
    }, [id, isEditing, setIsEditing, onSelectBlock])

    const onEnterBlock = useCallback(() => setIsEditing(true), [setIsEditing])

    const { onSelect } = useBlockSelection({
        id,
        blockElement: blockElement.current,
        onSelectBlock,
        isSelected,
        isInputFocused,
    })

    const { onKeyDown } = useBlockShortcuts({
        id,
        isMacPlatform,
        onMoveBlockSelection,
        onEnterBlock,
        onDeleteBlock,
        onRunBlock: runBlock,
        onMoveBlock,
        onDuplicateBlock,
    })

    useEffect(() => {
        if (isEditing) {
            editor?.focus()
        } else {
            blockElement.current?.focus()
        }
    }, [isEditing, editor])

    const modifierKeyLabel = isMacPlatform ? '⌘' : 'Ctrl'
    const commonMenuActions = useCommonBlockMenuActions({
        modifierKeyLabel,
        isInputFocused,
        onDeleteBlock,
        onMoveBlock,
        onDuplicateBlock,
    })
    const menuActions = useMemo(
        () =>
            [
                isEditing
                    ? {
                          label: 'Render',
                          icon: <PlayCircleOutlineIcon className="icon-inline" />,
                          onClick: runBlock,
                          keyboardShortcutLabel: `${modifierKeyLabel} + ↵`,
                      }
                    : {
                          label: 'Edit',
                          icon: <PencilIcon className="icon-inline" />,
                          onClick: onEnterBlock,
                          keyboardShortcutLabel: '↵',
                      },
            ].concat(commonMenuActions),
        [isEditing, modifierKeyLabel, runBlock, onEnterBlock, commonMenuActions]
    )

    const blockMenu = isSelected && <SearchNotebookBlockMenu id={id} actions={menuActions} />

    if (!isEditing) {
        return (
            <div className={classNames('block-wrapper', blockStyles.blockWrapper)}>
                {/* eslint-disable-next-line jsx-a11y/no-static-element-interactions */}
                <div
                    className={classNames(blockStyles.block, isSelected && blockStyles.selected, styles.outputWrapper)}
                    onClick={onSelect}
                    onFocus={onSelect}
                    onDoubleClick={onDoubleClick}
                    onKeyDown={onKeyDown}
                    // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
                    tabIndex={0}
                    // eslint-disable-next-line jsx-a11y/aria-role
                    role="notebook-block"
                    aria-label="Notebook block"
                    data-block-id={id}
                    ref={blockElement}
                >
                    <div className={styles.output}>
                        <Markdown dangerousInnerHTML={output ?? ''} />
                    </div>
                </div>
                {blockMenu}
            </div>
        )
    }

    return (
        <div className={classNames('block-wrapper', blockStyles.blockWrapper)}>
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
                // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
                tabIndex={0}
                // eslint-disable-next-line jsx-a11y/aria-role
                role="notebook-block"
                aria-label="Notebook block"
                data-block-id={id}
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
