import classNames from 'classnames'
import * as Monaco from 'monaco-editor'
import React, { useState, useCallback } from 'react'
import { useLocation } from 'react-router'
import { Observable, of } from 'rxjs'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { MonacoEditor } from '@sourcegraph/web/src/components/MonacoEditor'

import { SOURCEGRAPH_SEARCH } from '../input/MonacoQueryInput'
import { StreamingSearchResultsList } from '../results/StreamingSearchResultsList'

import blockStyles from './SearchNotebookBlock.module.scss'
import styles from './SearchNotebookQueryBlock.module.scss'
import { useBlockMonacoInput } from './useBlockMonacoEditor'

import { BlockProps, QueryBlock } from '.'

interface SearchNotebookQueryBlockProps
    extends BlockProps,
        Omit<QueryBlock, 'type'>,
        ThemeProps,
        SettingsCascadeProps,
        TelemetryProps {
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
    isLightTheme,
    telemetryService,
    settingsCascade,
    isSelected,
    fetchHighlightedFileLineRanges,
    onRunBlock,
    onBlockInputChange,
    onSelectBlock,
}) => {
    const [editor, setEditor] = useState<Monaco.editor.IStandaloneCodeEditor>()

    const { isInputFocused } = useBlockMonacoInput({
        editor,
        id,
        onRunBlock,
        onBlockInputChange,
        onSelectBlock,
    })

    const searchResults = useObservable(output ?? of(undefined))
    const location = useLocation()

    const onSelect = useCallback(() => {
        onSelectBlock(id)
    }, [id, onSelectBlock])

    return (
        <div
            className={classNames(
                blockStyles.block,
                styles.block,
                isSelected && !isInputFocused && blockStyles.selected,
                isSelected && isInputFocused && blockStyles.selectedNotFocused
            )}
            onClick={onSelect}
            role="presentation"
            data-block-id={id}
        >
            <div className={classNames(blockStyles.monacoWrapper, isInputFocused && blockStyles.selected)}>
                <MonacoEditor
                    language={SOURCEGRAPH_SEARCH}
                    value={input}
                    height={75}
                    isLightTheme={isLightTheme}
                    editorWillMount={() => {}}
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
