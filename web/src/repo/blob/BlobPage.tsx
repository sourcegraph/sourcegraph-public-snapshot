import ErrorIcon from '@sourcegraph/icons/lib/Error'
import * as H from 'history'
import { isEqual, pick, upperFirst } from 'lodash'
import * as React from 'react'
import { combineLatest, Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'
import { ExtensionsProps, ModeSpec } from '../../backend/features'
import { gql, queryGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { createAggregateError, ErrorLike, isErrorLike } from '../../util/errors'
import { memoizeObservable } from '../../util/memoize'
import { lprToRange, parseHash } from '../../util/url'
import { AbsoluteRepoFile, makeRepoURI, ParsedRepoURI } from '../index'
import { RepoHeaderActionPortal } from '../RepoHeaderActionPortal'
import { ToggleLineWrap } from './actions/ToggleLineWrap'
import { ToggleRenderedFileMode } from './actions/ToggleRenderedFileMode'
import { Blob2 } from './Blob2'
import { BlobPanel } from './panel/BlobPanel'
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
    }): Observable<GQL.IGitBlob> =>
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

interface Props extends AbsoluteRepoFile, ModeSpec, ExtensionsProps {
    location: H.Location
    history: H.History
    isLightTheme: boolean
    repoID: GQL.ID
}

interface State {
    wrapCode: boolean

    /**
     * The blob data or error that happened.
     * undefined while loading.
     */
    blobOrError?: GQL.IGitBlob | ErrorLike
}

export class BlobPage extends React.PureComponent<Props, State> {
    private propsUpdates = new Subject<Props>()
    private extendHighlightingTimeoutClicks = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = {
            wrapCode: ToggleLineWrap.getValue(),
        }
    }

    private logViewEvent(): void {
        eventLogger.logViewEvent('Blob', { fileShown: true })
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
                this.extendHighlightingTimeoutClicks.pipe(
                    mapTo(true),
                    startWith(false)
                )
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
        // renderAs is renderMode but with undefined mapped to the actual mode.
        const renderAs = renderMode || (this.state.blobOrError.richHTML ? 'rendered' : 'code')

        return (
            <>
                <PageTitle title={this.getPageTitle()} />
                {renderAs === 'code' && (
                    <RepoHeaderActionPortal
                        position="right"
                        priority={99}
                        element={<ToggleLineWrap key="toggle-line-wrap" onDidUpdate={this.onDidUpdateLineWrap} />}
                    />
                )}
                {this.state.blobOrError.richHTML && (
                    <RepoHeaderActionPortal
                        position="right"
                        priority={100}
                        element={
                            <ToggleRenderedFileMode
                                key="toggle-rendered-file-mode"
                                mode={renderMode || 'rendered'}
                                location={this.props.location}
                            />
                        }
                    />
                )}
                {this.state.blobOrError.richHTML &&
                    renderAs === 'rendered' && (
                        <RenderedFile
                            dangerousInnerHTML={this.state.blobOrError.richHTML}
                            location={this.props.location}
                        />
                    )}
                {renderAs === 'code' &&
                    !this.state.blobOrError.highlight.aborted && (
                        <>
                            <Blob2
                                className="blob-page__blob"
                                repoPath={this.props.repoPath}
                                commitID={this.props.commitID}
                                filePath={this.props.filePath}
                                html={this.state.blobOrError.highlight.html}
                                rev={this.props.rev}
                                mode={this.props.mode}
                                extensions={this.props.extensions}
                                wrapCode={this.state.wrapCode}
                                renderMode={renderMode}
                                location={this.props.location}
                                history={this.props.history}
                            />
                            <BlobPanel
                                {...this.props}
                                repoID={this.props.repoID}
                                repoPath={this.props.repoPath}
                                commitID={this.props.commitID}
                                extensions={this.props.extensions}
                                position={
                                    lprToRange(parseHash(this.props.location.hash))
                                        ? lprToRange(parseHash(this.props.location.hash))!.start
                                        : undefined
                                }
                            />
                        </>
                    )}
                {!this.state.blobOrError.richHTML &&
                    this.state.blobOrError.highlight.aborted && (
                        <div className="blob-page__aborted">
                            <div className="alert alert-info">
                                Syntax-highlighting this file took too long. &nbsp;
                                <button
                                    onClick={this.onExtendHighlightingTimeoutClick}
                                    className="btn btn-sm btn-primary"
                                >
                                    Try again
                                </button>
                            </div>
                        </div>
                    )}
            </>
        )
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
