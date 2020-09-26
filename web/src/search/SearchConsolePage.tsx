import React, { useCallback, useEffect, useMemo, useState } from 'react'
import * as H from 'history'
import { ThemeProps } from '../../../shared/src/theme'
import { PageTitle } from '../components/PageTitle'
import * as Monaco from 'monaco-editor'
import { MonacoEditor } from '../components/MonacoEditor'
import * as GQL from '../../../shared/src/graphql/schema'
import { SearchResultsList, SearchResultsListProps } from './results/SearchResultsList'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { ErrorLike } from '../../../shared/src/util/errors'
import { addSouregraphSearchCodeIntelligence } from './input/MonacoQueryInput'
import { BehaviorSubject, concat, of } from 'rxjs'
import { useEventObservable } from '../../../shared/src/util/useObservable'
import { switchMap, switchMapTo } from 'rxjs/operators'
import { search } from './backend'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { Omit } from 'utility-types'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { parseSearchURL } from '.'

interface SearchConsolePageProps
    extends ThemeProps,
        SettingsCascadeProps,
        Omit<SearchResultsListProps, 'extensionsController'>,
        ExtensionsControllerProps<'executeCommand' | 'services' | 'extHostAPI'> {
    globbing: boolean
    history: H.History
    location: H.Location
}

export const SearchConsolePage: React.FunctionComponent<SearchConsolePageProps> = props => {
    const searchQueries = useMemo(() => new BehaviorSubject<string>(parseSearchURL(location.search).query || ''), [])
    const [nextSearch, resultsOrError] = useEventObservable<'loading' | GQL.ISearchResults | ErrorLike>(
        useCallback(
            searchRequests =>
                searchRequests.pipe(
                    switchMapTo(searchQueries),
                    switchMap(query =>
                        concat(
                            of('loading' as const),
                            search(query, 'V2', props.patternType, undefined, props.extensionsController.extHostAPI)
                        )
                    )
                ),
            [searchQueries, props.patternType, props.extensionsController]
        )
    )
    const options: Monaco.editor.IEditorOptions = {
        readOnly: false,
        minimap: {
            enabled: false,
        },
        lineNumbers: 'off',
        fontSize: 14,
        glyphMargin: false,
        overviewRulerBorder: false,
        rulers: [],
        overviewRulerLanes: 0,
        wordBasedSuggestions: false,
        quickSuggestions: false,
        fixedOverflowWidgets: true,
        renderLineHighlight: 'none',
        contextmenu: false,
        links: false,
        // Display the cursor as a 1px line.
        cursorStyle: 'line',
        cursorWidth: 1,
    }
    const [monacoInstance, setMonacoInstance] = useState<typeof Monaco>()
    useEffect(() => {
        if (!monacoInstance) {
            return
        }
        const subscription = addSouregraphSearchCodeIntelligence(
            monacoInstance,
            searchQueries,
            of(props.patternType),
            of(props.globbing)
        )
        return () => subscription.unsubscribe()
    }, [monacoInstance, searchQueries, props.patternType, props.globbing])
    const [editorInstance, setEditorInstance] = useState<Monaco.editor.IStandaloneCodeEditor>()
    useEffect(() => {
        if (!editorInstance) {
            return
        }
        const disposable = editorInstance.onDidChangeModelContent(() => {
            const query = editorInstance.getValue()
            searchQueries.next(query)
            props.history.push('/search/console?q=' + encodeURI(query))
        })
        return () => disposable.dispose()
    }, [editorInstance, searchQueries, props.history])

    const calculateCount = (): number => {
        // This function can only get called if the results were successfully loaded,
        // so casting is the right thing to do here
        const results = resultsOrError as GQL.ISearchResults

        const parameters = new URLSearchParams(location.search)
        const query = parameters.get('q') || ''

        if (/count:(\d+)/.test(query)) {
            return Math.max(results.matchCount * 2, 1000)
        }
        return Math.max(results.matchCount * 2 || 0, 1000)
    }
    const showMoreResults = (): void => {
        // Requery with an increased max result count.
        if (!editorInstance) {
            return
        }
        let query = editorInstance.getValue()

        const count = calculateCount()
        if (/count:(\d+)/.test(query)) {
            console.log(`count:${count}`)
            query = query.replace(/count:\d+/g, '').trim() + ` count:${count}`
        } else {
            query = `${query} count:${count}`
        }
        editorInstance.setValue(query)
        props.history.push('/search/console?q=' + encodeURI(query))
    }

    return (
        <div className="w-100 p-2">
            <PageTitle title="Search console" />
            <div className="d-flex">
                <div className="flex-1 p-1">
                    <div className="mb-1 d-flex align-items-center justify-content-between">
                        <div />
                        <button className="btn btn-primary" type="button" onClick={nextSearch}>
                            Search
                        </button>
                    </div>
                    <MonacoEditor
                        {...props}
                        language="sourcegraphSearch"
                        options={options}
                        height={600}
                        editorWillMount={setMonacoInstance}
                        onEditorCreated={setEditorInstance}
                        value={searchQueries.value}
                    />
                </div>
                <div className="flex-1 p-1 search-console-page__results">
                    {resultsOrError &&
                        (resultsOrError === 'loading' ? (
                            <LoadingSpinner />
                        ) : (
                            <SearchResultsList
                                {...props}
                                resultsOrError={resultsOrError}
                                onShowMoreResultsClick={showMoreResults}
                            />
                        ))}
                </div>
            </div>
        </div>
    )
}
