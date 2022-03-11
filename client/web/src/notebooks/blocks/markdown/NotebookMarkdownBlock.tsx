import classNames from 'classnames'
import { noop } from 'lodash'
import PencilIcon from 'mdi-react/PencilIcon'
import PlayCircleOutlineIcon from 'mdi-react/PlayCircleOutlineIcon'
import * as Monaco from 'monaco-editor'
import React, { useState, useCallback, useEffect, useMemo } from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { MonacoEditor } from '@sourcegraph/shared/src/components/MonacoEditor'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BlockProps, MarkdownBlock } from '../..'
import { BlockMenuAction } from '../menu/NotebookBlockMenu'
import { useCommonBlockMenuActions } from '../menu/useCommonBlockMenuActions'
import { NotebookBlock } from '../NotebookBlock'
import blockStyles from '../NotebookBlock.module.scss'
import { useModifierKeyLabel } from '../useModifierKeyLabel'
import { MONACO_BLOCK_INPUT_OPTIONS, useMonacoBlockInput } from '../useMonacoBlockInput'

import styles from './NotebookMarkdownBlock.module.scss'

interface NotebookMarkdownBlockProps extends BlockProps<MarkdownBlock>, ThemeProps {}

export const NotebookMarkdownBlock: React.FunctionComponent<NotebookMarkdownBlockProps> = ({
    id,
    input,
    output,
    isSelected,
    isLightTheme,
    isReadOnly,
    onBlockInputChange,
    onRunBlock,
    onSelectBlock,
    ...props
}) => {
    const [isEditing, setIsEditing] = useState(!isReadOnly && input.length === 0)
    const [editor, setEditor] = useState<Monaco.editor.IStandaloneCodeEditor>()

    const runBlock = useCallback(
        (id: string) => {
            onRunBlock(id)
            setIsEditing(false)
        },
        [onRunBlock, setIsEditing]
    )

    const onInputChange = useCallback((input: string) => onBlockInputChange(id, { type: 'md', input }), [
        id,
        onBlockInputChange,
    ])

    const { isInputFocused } = useMonacoBlockInput({
        editor,
        id,
        ...props,
        onInputChange,
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

    useEffect(() => {
        if (isEditing) {
            editor?.focus()
        }
    }, [isEditing, editor])

    const commonMenuActions = useCommonBlockMenuActions({
        isInputFocused,
        isReadOnly,
        ...props,
    })

    const modifierKeyLabel = useModifierKeyLabel()
    const menuActions = useMemo(() => {
        const action: BlockMenuAction[] = [
            isEditing
                ? {
                      type: 'button',
                      label: 'Render',
                      icon: <PlayCircleOutlineIcon className="icon-inline" />,
                      onClick: runBlock,
                      keyboardShortcutLabel: `${modifierKeyLabel} + ↵`,
                  }
                : {
                      type: 'button',
                      label: 'Edit',
                      icon: <PencilIcon className="icon-inline" />,
                      onClick: onEnterBlock,
                      keyboardShortcutLabel: '↵',
                  },
        ]
        return action.concat(commonMenuActions)
    }, [isEditing, modifierKeyLabel, runBlock, onEnterBlock, commonMenuActions])

    const notebookBlockProps = useMemo(
        () => ({
            id,
            isInputFocused,
            onEnterBlock,
            isReadOnly,
            isSelected,
            onRunBlock,
            onBlockInputChange,
            onSelectBlock,
            actions: isSelected ? menuActions : [],
            'aria-label': 'Notebook markdown block',
            ...props,
        }),
        [
            id,
            isInputFocused,
            isReadOnly,
            isSelected,
            menuActions,
            onBlockInputChange,
            onEnterBlock,
            onRunBlock,
            onSelectBlock,
            props,
        ]
    )

    if (!isEditing) {
        return (
            <NotebookBlock {...notebookBlockProps} onDoubleClick={onDoubleClick}>
                <div className={styles.output} data-testid="output">
                    <Markdown className={styles.markdown} dangerousInnerHTML={output ?? ''} />
                </div>
            </NotebookBlock>
        )
    }

    return (
        <NotebookBlock
            className={classNames(styles.input, (isInputFocused || isSelected) && blockStyles.selected)}
            {...notebookBlockProps}
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
        </NotebookBlock>
    )
}
