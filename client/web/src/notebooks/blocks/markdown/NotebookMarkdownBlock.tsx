import React, { useState, useCallback, useMemo } from 'react'

import { insertTab } from '@codemirror/commands'
import { tags, HighlightStyle, classHighlightStyle } from '@codemirror/highlight'
import { markdown, markdownLanguage } from '@codemirror/lang-markdown'
import { Extension } from '@codemirror/state'
import { EditorView, keymap } from '@codemirror/view'
import classNames from 'classnames'
import PencilIcon from 'mdi-react/PencilIcon'
import PlayCircleOutlineIcon from 'mdi-react/PlayCircleOutlineIcon'

import { useCodeMirror } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Icon } from '@sourcegraph/wildcard'

import { BlockProps, MarkdownBlock } from '../..'
import { BlockMenuAction } from '../menu/NotebookBlockMenu'
import { useCommonBlockMenuActions } from '../menu/useCommonBlockMenuActions'
import { NotebookBlock } from '../NotebookBlock'
import { focusLastPositionInMonacoEditor, useFocusMonacoEditorOnMount } from '../useFocusMonacoEditorOnMount'
import { useIsBlockInputFocused } from '../useIsBlockInputFocused'
import { useModifierKeyLabel } from '../useModifierKeyLabel'

import blockStyles from '../NotebookBlock.module.scss'
import styles from './NotebookMarkdownBlock.module.scss'

const markdownExtension: Extension[] = [
    keymap.of([
        {
            // TODO: Maybe be smart about this an indent the line if it's the start
            // of a list item.
            // See also https://codemirror.net/6/examples/tab/ regarding how to
            // handle tab properly.
            key: 'Tab',
            run: insertTab,
        },
    ]),
    markdown({ base: markdownLanguage }),
    classHighlightStyle,
    HighlightStyle.define([
        { tag: tags.monospace, class: styles.mkdCode },
        { tag: tags.url, class: styles.mkdCode },
    ]),
]

interface NotebookMarkdownBlockProps extends BlockProps<MarkdownBlock>, ThemeProps {}

export const NotebookMarkdownBlock: React.FunctionComponent<NotebookMarkdownBlockProps> = React.memo(
    ({ id, input, output, isSelected, isLightTheme, isReadOnly, onBlockInputChange, onRunBlock, ...props }) => {
    const [isEditing, setIsEditing] = useState(!isReadOnly && input.initialFocusInput)
    const [container, setContainer] = useState<HTMLDivElement | null>(null)

    const runBlock = useCallback(() => {
        console.log('runBlock')
        onRunBlock(id)
        setIsEditing(false)
        return true
    }, [id, onRunBlock, setIsEditing])

    const onInputChange = useCallback((text: string) => onBlockInputChange(id, { type: 'md', input: { text } }), [
        id,
        onBlockInputChange,
    ])

    const extensions: Extension[] = useMemo(
        () => [
            keymap.of([
                {
                    key: 'Mod-Enter',
                    run: runBlock,
                },
            ]),
            EditorView.updateListener.of(update => {
                if (update.docChanged) {
                    console.log(update.state.sliceDoc())
                    onInputChange(update.state.sliceDoc())
                }
            }),
            markdownExtension,
        ],
        [runBlock, onInputChange]
    )

    const editor = useCodeMirror(container, input.text, extensions)

    const editMarkdown = useCallback(() => {
        if (!isReadOnly) {
            setIsEditing(true)
        }
    }, [isReadOnly, setIsEditing])

    // useFocusMonacoEditorOnMount({ editor, isEditing })

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
                      onClick: editMarkdown,
                      keyboardShortcutLabel: '↵',
                  },
        ]
        return action.concat(commonMenuActions)
    }, [isEditing, modifierKeyLabel, runBlock, editMarkdown, commonMenuActions])

    const focusInput = useCallback(() => {
        editor?.dispatch({
            selection: {anchor: editor.state.doc.length},
            scrollIntoView: true,
        })
    }, [editor])

    const notebookBlockProps = useMemo(
        () => ({
                id,
                isReadOnly,
                isSelected,
                onRunBlock,
                onBlockInputChange,
                actions: isSelected && !isReadOnly ? menuActions : [],
                'aria-label': 'Notebook markdown block',
                isInputVisible: isEditing,
                setIsInputVisible: setIsEditing,
                focusInput,
                ...props,
            }),
            [id, isEditing, isReadOnly, isSelected, menuActions, onBlockInputChange, onRunBlock, focusInput, props]
        )

        const isInputFocused = useIsBlockInputFocused(id)

        if (!isEditing) {
            return (
                <NotebookBlock {...notebookBlockProps} onDoubleClick={editMarkdown}>
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
            <div ref={setContainer} className={blockStyles.monacoWrapper} />
        </NotebookBlock>
    )
})
