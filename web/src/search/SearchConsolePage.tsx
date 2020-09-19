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
import { Toggles } from './input/toggles/Toggles'
import { addSouregraphSearchCodeIntelligence } from './input/MonacoQueryInput'
import { BehaviorSubject, concat, of } from 'rxjs'
import { useEventObservable } from '../../../shared/src/util/useObservable'
import { switchMap, switchMapTo } from 'rxjs/operators'
import { search } from './backend'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { Omit } from 'utility-types'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

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
    const searchQueries = useMemo(() => new BehaviorSubject<string>(''), [])
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
            searchQueries.next(editorInstance.getValue())
        })
        return () => disposable.dispose()
    }, [editorInstance, searchQueries])
    return (
        <div className="w-100 p-2">
            <PageTitle title="Search console" />
            <div className="d-flex">
                <div className="flex-1 p-1">
                    <div className="mb-1 d-flex align-items-center justify-content-between">
                        <Toggles {...props} copyQueryButton={false} navbarSearchQuery={searchQueries.value} />
                        <button className="btn btn-primary" type="button" onClick={nextSearch}>
                            Search
                        </button>
                    </div>
                    <MonacoEditor
                        {...props}
                        language="sourcegraphSearch"
                        options={options}
                        height={400}
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
                            <SearchResultsList {...props} resultsOrError={resultsOrError} />
                        ))}
                </div>
            </div>
        </div>
    )
}
