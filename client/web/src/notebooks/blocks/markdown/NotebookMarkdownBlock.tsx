import React, { useState, useCallback, useMemo, useEffect, useRef } from 'react'

import { defaultKeymap, indentWithTab, history, historyKeymap } from '@codemirror/commands'
import { markdown, markdownLanguage } from '@codemirror/lang-markdown'
import { indentUnit, HighlightStyle, syntaxHighlighting } from '@codemirror/language'
import type { Extension } from '@codemirror/state'
import { EditorView, keymap } from '@codemirror/view'
import { classHighlighter, tags } from '@lezer/highlight'
import { mdiPlayCircleOutline, mdiPencil } from '@mdi/js'
import classNames from 'classnames'

import { changeListener } from '@sourcegraph/branded'
import { CodeMirrorEditor, type Editor, editorHeight } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { Icon, Markdown } from '@sourcegraph/wildcard'

import type { BlockProps, MarkdownBlock } from '../..'
import type { BlockMenuAction } from '../menu/NotebookBlockMenu'
import { useCommonBlockMenuActions } from '../menu/useCommonBlockMenuActions'
import { NotebookBlock } from '../NotebookBlock'
import { useIsBlockInputFocused } from '../useIsBlockInputFocused'
import { useModifierKeyLabel } from '../useModifierKeyLabel'

import blockStyles from '../NotebookBlock.module.scss'
import styles from './NotebookMarkdownBlock.module.scss'

const staticExtensions: Extension[] = [
    history(),
    keymap.of([
        // Insert a soft tab if the cursor is not at the beginning of the line
        // or a list item.
        {
            key: 'Tab',
            run: view => {
                const { main } = view.state.selection

                // If text is actually selected, fall back to indentation
                // instead
                if (main.from !== main.to) {
                    return false
                }
                const currentLine = view.state.doc.lineAt(main.to)
                if (/^\s*((-|\*)\s*)?$/.test(view.state.sliceDoc(currentLine.from, main.to))) {
                    // We could be smarter about this and actually inspect the
                    // syntax tree, but maybe this is good enough?
                    return false
                }

                // Insert a soft tab
                const indent = view.state.facet(indentUnit)
                view.dispatch({
                    changes: {
                        from: main.to,
                        insert: indent,
                    },
                    selection: { anchor: main.to + indent.length },
                })
                return true
            },
        },
        {
            key: 'Escape',
            run: view => {
                view.contentDOM.blur()
                return true
            },
        },
        indentWithTab,
    ]),
    keymap.of(historyKeymap),
    keymap.of(defaultKeymap),
    EditorView.lineWrapping,
    markdown({ base: markdownLanguage }),
    syntaxHighlighting(classHighlighter),
    syntaxHighlighting(
        HighlightStyle.define([
            { tag: tags.monospace, class: styles.markdownCode },
            { tag: tags.url, class: styles.markdownCode },
        ])
    ),
    editorHeight({ maxHeight: '60rem' }),
]

interface NotebookMarkdownBlockProps extends Omit<BlockProps<MarkdownBlock>, 'patternType'> {
    isEmbedded?: boolean
}

export const NotebookMarkdownBlock: React.FunctionComponent<React.PropsWithChildren<NotebookMarkdownBlockProps>> =
    React.memo(
        ({
            id,
            input,
            output,
            isSelected,
            isReadOnly,
            isEmbedded,
            onBlockInputChange,
            onRunBlock,
            onNewBlock,
            ...props
        }) => {
            const [isEditing, setIsEditing] = useState(!isReadOnly && input.initialFocusInput)
            const editorRef = useRef<Editor | null>(null)

            const runBlock = useCallback(() => {
                onRunBlock(id)
                setIsEditing(false)
                return true
            }, [id, onRunBlock, setIsEditing])

            const newBlock = useCallback(() => {
                onRunBlock(id)
                onNewBlock(id)
                setIsEditing(false)
                return true
            }, [id, onRunBlock, onNewBlock, setIsEditing])

            const onInputChange = useCallback(
                (text: string) => onBlockInputChange(id, { type: 'md', input: { text } }),
                [id, onBlockInputChange]
            )

            const editorExtensions: Extension[] = useMemo(
                () => [
                    keymap.of([
                        {
                            key: 'Mod-Enter',
                            run: runBlock,
                        },
                        {
                            key: 'Mod-Shift-Enter',
                            run: newBlock,
                        },
                    ]),
                    changeListener(onInputChange),
                    staticExtensions,
                ],
                [runBlock, onInputChange, newBlock]
            )

            const editMarkdown = useCallback(() => {
                if (!isReadOnly) {
                    setIsEditing(true)
                }
            }, [isReadOnly, setIsEditing])

            useEffect(() => {
                editorRef.current?.focus()
            }, [isEditing])

            const commonMenuActions = useCommonBlockMenuActions({ id, isReadOnly, onNewBlock, ...props })

            const modifierKeyLabel = useModifierKeyLabel()
            const menuActions = useMemo(() => {
                const action: BlockMenuAction[] = [
                    isEditing
                        ? {
                              type: 'button',
                              label: 'Render',
                              icon: <Icon aria-hidden={true} svgPath={mdiPlayCircleOutline} />,
                              onClick: runBlock,
                              keyboardShortcutLabel: `${modifierKeyLabel} + ↵`,
                          }
                        : {
                              type: 'button',
                              label: 'Edit',
                              icon: <Icon aria-hidden={true} svgPath={mdiPencil} />,
                              onClick: editMarkdown,
                              keyboardShortcutLabel: '↵',
                          },
                ]
                return action.concat(commonMenuActions)
            }, [isEditing, modifierKeyLabel, runBlock, editMarkdown, commonMenuActions])

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
                    focusInput: () => editorRef.current?.focus(),
                    ...props,
                }),
                [id, isEditing, isReadOnly, isSelected, menuActions, onBlockInputChange, onRunBlock, editorRef, props]
            )

            const isInputFocused = useIsBlockInputFocused(id)

            if (!isEditing) {
                return (
                    <NotebookBlock {...notebookBlockProps} onDoubleClick={editMarkdown}>
                        <div
                            className={classNames(styles.output, isEmbedded && styles.isEmbedded)}
                            data-testid="output"
                        >
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
                    <CodeMirrorEditor ref={editorRef} value={input.text} extensions={editorExtensions} />
                </NotebookBlock>
            )
        }
    )
NotebookMarkdownBlock.displayName = 'NotebookMarkdownBlock'
