import React, { useState, useCallback, useEffect, useMemo } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'
import PencilIcon from 'mdi-react/PencilIcon'
import PlayCircleOutlineIcon from 'mdi-react/PlayCircleOutlineIcon'
import * as Monaco from 'monaco-editor'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { MonacoEditor } from '@sourcegraph/shared/src/components/MonacoEditor'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Icon } from '@sourcegraph/wildcard'

import { BlockProps, MarkdownBlock } from '../..'
import { BlockMenuAction } from '../menu/NotebookBlockMenu'
import { useCommonBlockMenuActions } from '../menu/useCommonBlockMenuActions'
import { NotebookBlock } from '../NotebookBlock'
import { useIsBlockInputFocused } from '../useIsBlockInputFocused'
import { useModifierKeyLabel } from '../useModifierKeyLabel'
import {
    focusLastPositionInMonacoEditor,
    MONACO_BLOCK_INPUT_OPTIONS,
    useMonacoBlockInput,
} from '../useMonacoBlockInput'

import blockStyles from '../NotebookBlock.module.scss'
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
    ...props
}) => {
    const [isEditing, setIsEditing] = useState(!isReadOnly && input.initialFocusInput)
    const [editor, setEditor] = useState<Monaco.editor.IStandaloneCodeEditor>()

    const runBlock = useCallback(
        (id: string) => {
            onRunBlock(id)
            setIsEditing(false)
        },
        [onRunBlock, setIsEditing]
    )

    const onInputChange = useCallback((text: string) => onBlockInputChange(id, { type: 'md', input: { text } }), [
        id,
        onBlockInputChange,
    ])

    useMonacoBlockInput({
        editor,
        id,
        tabMovesFocus: false,
        ...props,
        onInputChange,
        onRunBlock: runBlock,
    })

    const onDoubleClick = useCallback(() => {
        if (isReadOnly) {
            return
        }
        if (!isEditing) {
            setIsEditing(true)
        }
    }, [isReadOnly, isEditing, setIsEditing])

    const onEnterBlock = useCallback(() => {
        if (isReadOnly) {
            return
        }
        setIsEditing(true)
    }, [isReadOnly, setIsEditing])

    useEffect(() => {
        if (isEditing && editor) {
            focusLastPositionInMonacoEditor(editor)
        }
    }, [isEditing, editor])

    const commonMenuActions = useCommonBlockMenuActions({ id, isReadOnly, ...props })

    const modifierKeyLabel = useModifierKeyLabel()
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

    const notebookBlockProps = useMemo(
        () => ({
            id,
            onEnterBlock,
            isReadOnly,
            isSelected,
            onRunBlock,
            onBlockInputChange,
            actions: isSelected && !isReadOnly ? menuActions : [],
            'aria-label': 'Notebook markdown block',
            ...props,
        }),
        [id, isReadOnly, isSelected, menuActions, onBlockInputChange, onEnterBlock, onRunBlock, props]
    )

    const isInputFocused = useIsBlockInputFocused(id)

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
                    value={input.text}
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
