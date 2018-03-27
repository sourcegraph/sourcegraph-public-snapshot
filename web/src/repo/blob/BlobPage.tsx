import ErrorIcon from '@sourcegraph/icons/lib/Error'
import * as H from 'history'
import isEqual from 'lodash/isEqual'
import pick from 'lodash/pick'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { combineLatest } from 'rxjs/observable/combineLatest'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { mapTo } from 'rxjs/operators/mapTo'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { gql, queryGraphQL } from '../../backend/graphql'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { createAggregateError, ErrorLike, isErrorLike } from '../../util/errors'
import { memoizeObservable } from '../../util/memoize'
import { parseHash } from '../../util/url'
import { makeRepoURI, ParsedRepoURI } from '../index'
import { RepoHeaderActionPortal } from '../RepoHeaderActionPortal'
import { OpenInEditorAction } from './actions/OpenInEditorAction'
import { ToggleLineWrap } from './actions/ToggleLineWrap'
import { ToggleRenderedFileMode } from './actions/ToggleRenderedFileMode'
import { Blob } from './Blob'
import { BlobPanel } from './BlobPanel'
import { RenderedFile } from './RenderedFile'

function fetchBlobCacheKey(parsed: ParsedRepoURI & { isLightTheme: boolean; disableTimeout: boolean }): string {
    return makeRepoURI(parsed) + parsed.isLightTheme + parsed.disableTimeout
}

const fetchBlob = memoizeObservable(
    (args: {
        repoPath: string
        commitID: string
        filePath: string
        isLightTheme: boolean
        disableTimeout: boolean
    }): Observable<GQL.IFile> =>
        queryGraphQL(
            gql`
                query Blob(
                    $repoPath: String!
                    $commitID: String!
                    $filePath: String!
                    $isLightTheme: Boolean!
                    $disableTimeout: Boolean!
                ) {
                    repository(uri: $repoPath) {
                        commit(rev: $commitID) {
                            file(path: $filePath) {
                                richHTML
                                highlight(disableTimeout: $disableTimeout, isLightTheme: $isLightTheme) {
                                    aborted
                                    html
                                }
                            }
                        }
                    }
                }
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.repository ||
                    !data.repository.commit ||
                    !data.repository.commit.file ||
                    !data.repository.commit.file.highlight
                ) {
                    throw createAggregateError(errors)
                }
                return data.repository.commit.file
            })
        ),
    fetchBlobCacheKey
)

interface Props {
    location: H.Location
    history: H.History
    isLightTheme: boolean
    repoPath: string
    rev: string | undefined
    commitID: string
    filePath: string
}

interface State {
    wrapCode: boolean

    /**
     * Whether to show the references panel.
     */
    showRefs: boolean

    /**
     * The blob data or error that happened.
     * undefined while loading.
     */
    blobOrError?: GQL.IFile | ErrorLike
}

export class BlobPage extends React.PureComponent<Props, State> {
    private propsUpdates = new Subject<Props>()
    private extendHighlightingTimeoutClicks = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = {
            wrapCode: ToggleLineWrap.getValue(),
            showRefs: parseHash(props.location.hash).modal === 'references',
        }
    }

    private logViewEvent(): void {
        eventLogger.logViewEvent('Blob', { fileShown: true, referencesShown: this.state.showRefs })
    }

    public componentDidMount(): void {
        this.logViewEvent()

        // Fetch repository revision.
        this.subscriptions.add(
            combineLatest(
                this.propsUpdates.pipe(
                    map(props => pick(props, 'repoPath', 'commitID', 'filePath', 'isLightTheme')),
                    distinctUntilChanged((a, b) => isEqual(a, b))
                ),
                this.extendHighlightingTimeoutClicks.pipe(mapTo(true), startWith(false))
            )
                .pipe(
                    tap(() => this.setState({ blobOrError: undefined })),
                    switchMap(([{ repoPath, commitID, filePath, isLightTheme }, extendHighlightingTimeout]) =>
                        fetchBlob({
                            repoPath,
                            commitID,
                            filePath,
                            isLightTheme,
                            disableTimeout: extendHighlightingTimeout,
                        }).pipe(
                            catchError(error => {
                                console.error(error)
                                return [error]
                            })
                        )
                    )
                )
                .subscribe(blobOrError => this.setState({ blobOrError }), err => console.error(err))
        )

        this.propsUpdates.next(this.props)
    }

    public componentWillReceiveProps(newProps: Props): void {
        this.propsUpdates.next(newProps)
        if (
            newProps.repoPath !== this.props.repoPath ||
            newProps.commitID !== this.props.commitID ||
            newProps.filePath !== this.props.filePath ||
            ToggleRenderedFileMode.getModeFromURL(newProps.location) !==
                ToggleRenderedFileMode.getModeFromURL(this.props.location)
        ) {
            this.logViewEvent()
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        if (!this.state.blobOrError) {
            // Render placeholder for layout before content is fetched.
            return <div className="blob-page__placeholder" />
        }

        if (isErrorLike(this.state.blobOrError)) {
            return <HeroPage icon={ErrorIcon} title="Error" subtitle={upperFirst(this.state.blobOrError.message)} />
        }

        const renderMode = ToggleRenderedFileMode.getModeFromURL(this.props.location)
        const hash = parseHash(this.props.location.hash)

        return [
            <PageTitle key="page-title" title={this.getPageTitle()} />,
            <RepoHeaderActionPortal
                position="right"
                key="open-in-editor"
                element={
                    <OpenInEditorAction
                        key="open-in-editor"
                        repoPath={this.props.repoPath}
                        commitID={this.props.commitID}
                        filePath={this.props.filePath}
                        location={this.props.location}
                    />
                }
            />,
            renderMode === 'code' && (
                <RepoHeaderActionPortal
                    position="right"
                    key="toggle-line-wrap"
                    element={<ToggleLineWrap key="toggle-line-wrap" onDidUpdate={this.onDidUpdateLineWrap} />}
                />
            ),
            this.state.blobOrError.richHTML && (
                <RepoHeaderActionPortal
                    key="toggle-rendered-file-mode"
                    position="right"
                    element={
                        <ToggleRenderedFileMode
                            key="toggle-rendered-file-mode"
                            mode={renderMode}
                            location={this.props.location}
                        />
                    }
                />
            ),
            this.state.blobOrError.richHTML &&
                renderMode === 'rendered' && (
                    <RenderedFile
                        key="rendered-file"
                        dangerousInnerHTML={this.state.blobOrError.richHTML}
                        location={this.props.location}
                    />
                ),
            (renderMode === 'code' || !this.state.blobOrError.richHTML) &&
                !this.state.blobOrError.highlight.aborted && (
                    <Blob
                        key="blob"
                        className="blob-page__blob"
                        repoPath={this.props.repoPath}
                        commitID={this.props.commitID}
                        filePath={this.props.filePath}
                        html={this.state.blobOrError.highlight.html}
                        rev={this.props.rev}
                        wrapCode={this.state.wrapCode}
                        renderMode={renderMode}
                        location={this.props.location}
                        history={this.props.history}
                    />
                ),
            !this.state.blobOrError.richHTML &&
                this.state.blobOrError.highlight.aborted && (
                    <div className="blob-page__aborted" key="aborted">
                        <div className="alert alert-info">
                            Syntax-highlighting this file took too long. &nbsp;
                            <button onClick={this.onExtendHighlightingTimeoutClick} className="btn btn-sm btn-primary">
                                Try again
                            </button>
                        </div>
                    </div>
                ),
            hash.modal === 'references' &&
                hash.line && (
                    <BlobPanel
                        key="blob-panel"
                        repoPath={this.props.repoPath}
                        rev={this.props.rev}
                        commitID={this.props.commitID}
                        filePath={this.props.filePath}
                        line={hash.line}
                        character={hash.character || 0}
                        modalMode={hash.modalMode}
                        isLightTheme={this.props.isLightTheme}
                        location={this.props.location}
                        history={this.props.history}
                    />
                ),
        ]
    }

    private onDidUpdateLineWrap = (value: boolean) => this.setState({ wrapCode: value })

    private onExtendHighlightingTimeoutClick = () => this.extendHighlightingTimeoutClicks.next()

    private getPageTitle(): string {
        const repoPathSplit = this.props.repoPath.split('/')
        const repoStr = repoPathSplit.length > 2 ? repoPathSplit.slice(1).join('/') : this.props.repoPath
        if (this.props.filePath) {
            const fileOrDir = this.props.filePath.split('/').pop()
            return `${fileOrDir} - ${repoStr}`
        }
        return `${repoStr}`
    }
}
