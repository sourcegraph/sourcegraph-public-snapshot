import classNames from 'classnames'
import * as Monaco from 'monaco-editor'
import React, { useState, useEffect, useMemo } from 'react'
import { useLocation } from 'react-router'
import { BehaviorSubject, Observable, of } from 'rxjs'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { MonacoEditor } from '@sourcegraph/web/src/components/MonacoEditor'

import { fetchSuggestions } from '../backend'
import { addSourcegraphSearchCodeIntelligence, SOURCEGRAPH_SEARCH } from '../input/MonacoQueryInput'
import { StreamingSearchResultsList } from '../results/StreamingSearchResultsList'

import styles from './SearchNotebookQueryBlock.module.scss'

import { BlockProps, QueryBlock } from '.'

interface SearchNotebookQueryBlockProps
    extends BlockProps,
        Omit<QueryBlock, 'type'>,
        ThemeProps,
        SettingsCascadeProps,
        TelemetryProps {
    globbing: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}

// Move to somewhere common to share with markdown block
const options: Monaco.editor.IStandaloneEditorConstructionOptions = {
    readOnly: false,
    lineNumbers: 'off',
    lineHeight: 16,
    // Match the query input's height for suggestion items line height.
    suggestLineHeight: 34,
    minimap: {
        enabled: false,
    },
    scrollbar: {
        vertical: 'hidden',
        horizontal: 'hidden',
    },
    glyphMargin: false,
    hover: { delay: 150 },
    lineDecorationsWidth: 0,
    lineNumbersMinChars: 0,
    overviewRulerBorder: false,
    folding: false,
    rulers: [],
    overviewRulerLanes: 0,
    wordBasedSuggestions: false,
    quickSuggestions: false,
    fixedOverflowWidgets: true,
    contextmenu: false,
    links: false,
    // Match our monospace/code style from code.scss
    fontFamily: 'sfmono-regular, consolas, menlo, dejavu sans mono, monospace',
    // Display the cursor as a 1px line.
    cursorStyle: 'line',
    cursorWidth: 1,
    automaticLayout: true,
}

// TODO: Use React.memo
export const SearchNotebookQueryBlock: React.FunctionComponent<SearchNotebookQueryBlockProps> = ({
    id,
    input,
    output,
    globbing,
    isLightTheme,
    telemetryService,
    settingsCascade,
    fetchHighlightedFileLineRanges,
    onRunBlock,
    onBlockInputChange,
}) => {
    const [monacoInstance, setMonacoInstance] = useState<typeof Monaco>()
    const [editor, setEditor] = useState<Monaco.editor.IStandaloneCodeEditor>()
    const searchQueryInput = useMemo(() => new BehaviorSubject<string>(input), [input])

    useEffect(() => {
        if (!editor) {
            return
        }
        const disposable = editor.addAction({
            id: 'render-on-cmd-enter',
            label: 'Submit search',
            keybindings: [Monaco.KeyMod.CtrlCmd | Monaco.KeyCode.Enter],
            run: () => {
                onRunBlock(id)
            },
        })
        return () => disposable.dispose()
    }, [editor, id, onRunBlock])

    useEffect(() => {
        if (!editor) {
            return
        }
        const disposable = editor.onDidChangeModelContent(() => {
            const value = editor.getValue()
            onBlockInputChange(id, value)
            searchQueryInput.next(value)
        })
        return () => disposable.dispose()
    }, [editor, id, searchQueryInput, onBlockInputChange])

    useEffect(() => {
        if (!monacoInstance) {
            return
        }
        const subscription = addSourcegraphSearchCodeIntelligence(monacoInstance, searchQueryInput, fetchSuggestions, {
            // TODO?
            patternType: SearchPatternType.literal,
            globbing,
            interpretComments: true,
        })
        return () => subscription.unsubscribe()
    }, [monacoInstance, searchQueryInput, globbing])

    useEffect(() => {
        if (!editor) {
            return
        }
        const disposable = editor.addAction({
            id: 'run-on-cmd-enter',
            label: 'Run query',
            keybindings: [Monaco.KeyMod.CtrlCmd | Monaco.KeyCode.Enter],
            run: () => {
                onRunBlock(id)
            },
        })
        return () => disposable.dispose()
    }, [editor, id, onRunBlock])

    const searchResults = useObservable(output ?? of(undefined))

    const location = useLocation()
    return (
        <div className={styles.block}>
            <div className={styles.monacoWrapper}>
                <MonacoEditor
                    language={SOURCEGRAPH_SEARCH}
                    value={input}
                    height={75}
                    isLightTheme={isLightTheme}
                    editorWillMount={setMonacoInstance}
                    onEditorCreated={setEditor}
                    options={options}
                    border={false}
                />
            </div>

            {searchResults && searchResults.state === 'loading' && (
                <div className={classNames('d-flex justify-content-center py-3', styles.results)}>
                    <LoadingSpinner />
                </div>
            )}
            {searchResults && searchResults.state !== 'loading' && (
                <div className={styles.results}>
                    <StreamingSearchResultsList
                        location={location}
                        allExpanded={false}
                        results={searchResults}
                        isLightTheme={isLightTheme}
                        fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                        telemetryService={telemetryService}
                        settingsCascade={settingsCascade}
                    />
                </div>
            )}
        </div>
    )
}
