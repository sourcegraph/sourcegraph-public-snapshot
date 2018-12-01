import { highlight } from 'highlight.js/lib/highlight'
import * as H from 'history'
import { castArray, isEqual } from 'lodash'
import marked from 'marked'
import * as React from 'react'
import { concat, merge, Observable, of, Subject, Subscription } from 'rxjs'
import {
    bufferTime,
    catchError,
    delay,
    distinctUntilChanged,
    map,
    scan,
    skip,
    startWith,
    switchMap,
    takeUntil,
} from 'rxjs/operators'
import { ContributableViewContainer } from '../../../../../shared/src/api/protocol'
import { Location, Position } from '../../../../../shared/src/api/protocol/plainTypes'
import { RepositoryIcon } from '../../../../../shared/src/components/icons' // TODO: Switch to mdi icon
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { FileLocations } from '../../../../../shared/src/panel/views/FileLocations'
import { FileLocationsTree } from '../../../../../shared/src/panel/views/FileLocationsTree'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { AbsoluteRepoFile, parseHash, PositionSpec } from '../../../../../shared/src/util/url'
import {
    getDefinition,
    getHover,
    getImplementations,
    getReferences,
    HoverMerged,
    ModeSpec,
} from '../../../backend/features'
import { isEmptyHover, LSPTextDocumentPositionParams } from '../../../backend/lsp'
import { isDiscussionsEnabled } from '../../../discussions'
import { eventLogger } from '../../../tracking/eventLogger'
import { fetchHighlightedFileLines } from '../../backend'
import { RepoHeaderContributionsLifecycleProps } from '../../RepoHeader'
import { RepoRevSidebarCommits } from '../../RepoRevSidebarCommits'
import { DiscussionsTree } from '../discussions/DiscussionsTree'
import { fetchExternalReferences } from '../references/backend'

interface Props
    extends AbsoluteRepoFile,
        Partial<PositionSpec>,
        ModeSpec,
        RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        PlatformContextProps,
        ExtensionsControllerProps {
    location: H.Location
    history: H.History
    repoID: GQL.ID
    repoPath: string
    commitID: string
    isLightTheme: boolean
    authenticatedUser: GQL.IUser | null
}

/** The subject (what the contextual information refers to). */
interface ContextSubject extends ModeSpec, PlatformContextProps {
    repoPath: string
    commitID: string
    filePath: string
    line: number
    character: number
}

export type BlobPanelTabID = 'info' | 'def' | 'references' | 'references:external' | 'discussions' | 'impl' | 'history'

function toSubject(props: Props): ContextSubject {
    const parsedHash = parseHash(props.location.hash)
    return {
        repoPath: props.repoPath,
        commitID: props.commitID,
        filePath: props.filePath,
        mode: props.mode,
        line: parsedHash.line || 1,
        character: parsedHash.character || 1,
        platformContext: props.platformContext,
    }
}

function subjectIsEqual(a: ContextSubject, b: ContextSubject & { line?: number; character?: number }): boolean {
    return (
        a &&
        b &&
        a.repoPath === b.repoPath &&
        a.commitID === b.commitID &&
        a.filePath === b.filePath &&
        a.mode === b.mode &&
        a.line === b.line &&
        a.character === b.character &&
        isEqual(a.platformContext, b.platformContext)
    )
}

const LOADING: 'loading' = 'loading'

type PanelSubject = AbsoluteRepoFile & ModeSpec & PlatformContextProps & { position?: Position }

/**
 * A panel on the blob page that displays contextual information.
 */
export class BlobPanel extends React.PureComponent<Props> {
    private componentUpdates = new Subject<Props>()
    private locationsUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const componentUpdates = this.componentUpdates.pipe(startWith(this.props))

        // Changes to the context subject, including upon the initial mount.
        const subjectChanges = componentUpdates.pipe(
            distinctUntilChanged<Props>((a, b) => subjectIsEqual(toSubject(a), toSubject(b)))
        )

        // Info (hover) panel view.
        this.subscriptions.add(
            this.props.extensionsController.services.views.registerProvider(
                { id: 'Info', container: ContributableViewContainer.Panel },
                subjectChanges.pipe(
                    switchMap((subject: PanelSubject) => {
                        const { position } = subject
                        if (
                            !position ||
                            position.character === 0 /* 1-indexed, so this means only line (not position) is selected */
                        ) {
                            return [undefined]
                        }
                        const result = getHover(
                            {
                                ...(subject as Pick<typeof subject, Exclude<keyof typeof subject, 'platformContext'>>),
                                position,
                            },
                            { extensionsController: this.props.extensionsController }
                        ).pipe(catchError(error => [asError(error) as ErrorLike]))
                        return merge(
                            result,
                            of(LOADING).pipe(
                                delay(150),
                                takeUntil(result)
                            ) // delay loading spinner to reduce jitter
                        ).pipe(
                            startWith(undefined) // clear old data immediately
                        )
                    }),
                    map((hoverOrError: undefined | null | HoverMerged | ErrorLike | typeof LOADING) => {
                        if (
                            hoverOrError &&
                            hoverOrError !== LOADING &&
                            !isErrorLike(hoverOrError) &&
                            !isEmptyHover(hoverOrError)
                        ) {
                            if (Array.isArray(hoverOrError.contents) && hoverOrError.contents.length >= 2) {
                                return {
                                    title: 'Info',
                                    content: '',
                                    locationProvider: null,
                                    reactElement: hoverOrError.contents.map((s, i) => (
                                        <div key={i} className="blob-panel__extra-item px-2 pt-1">
                                            {renderHoverContents(s)}
                                        </div>
                                    )),
                                }
                            }
                        }
                        return null
                    })
                )
            )
        )

        // Definition panel view.
        this.subscriptions.add(
            this.props.extensionsController.services.views.registerProvider(
                { id: 'def', container: ContributableViewContainer.Panel },
                subjectChanges.pipe(
                    map((subject: PanelSubject) => ({
                        title: 'Definition',
                        content: '',
                        locationProvider: null,
                        reactElement: (
                            <FileLocations
                                className="panel__tabs-content"
                                query={this.queryDefinition}
                                inputRepo={subject.repoPath}
                                inputRevision={subject.rev}
                                // tslint:disable-next-line:jsx-no-lambda
                                onSelect={() => this.onSelectLocation('def')}
                                icon={RepositoryIcon}
                                pluralNoun="definitions"
                                isLightTheme={this.props.isLightTheme}
                                fetchHighlightedFileLines={fetchHighlightedFileLines}
                            />
                        ),
                    }))
                )
            )
        )

        // References panel view.
        this.subscriptions.add(
            this.props.extensionsController.services.views.registerProvider(
                { id: 'references', container: ContributableViewContainer.Panel },
                subjectChanges.pipe(
                    map((subject: PanelSubject) => ({
                        title: 'References',
                        content: '',
                        locationProvider: null,
                        reactElement: (
                            <FileLocations
                                className="panel__tabs-content"
                                query={this.queryReferencesLocal}
                                inputRepo={subject.repoPath}
                                inputRevision={subject.rev}
                                // tslint:disable-next-line:jsx-no-lambda
                                onSelect={() => this.onSelectLocation('references')}
                                icon={RepositoryIcon}
                                pluralNoun="local references"
                                isLightTheme={this.props.isLightTheme}
                                fetchHighlightedFileLines={fetchHighlightedFileLines}
                            />
                        ),
                    }))
                )
            )
        )

        // External references panel view.
        this.subscriptions.add(
            this.props.extensionsController.services.views.registerProvider(
                { id: 'references:external', container: ContributableViewContainer.Panel },
                subjectChanges.pipe(
                    map((subject: PanelSubject) => ({
                        title: 'External references',
                        content: '',
                        locationProvider: null,
                        reactElement: (
                            <FileLocationsTree
                                className="panel__tabs-content"
                                query={this.queryReferencesExternal}
                                // tslint:disable-next-line:jsx-no-lambda
                                onSelectLocation={() => this.onSelectLocation('references:external')}
                                icon={RepositoryIcon}
                                pluralNoun="external references"
                                isLightTheme={this.props.isLightTheme}
                                location={this.props.location}
                                fetchHighlightedFileLines={fetchHighlightedFileLines}
                            />
                        ),
                    }))
                )
            )
        )

        // Implementations panel view.
        this.subscriptions.add(
            this.props.extensionsController.services.views.registerProvider(
                { id: 'impl', container: ContributableViewContainer.Panel },
                subjectChanges.pipe(
                    map((subject: PanelSubject) => ({
                        title: 'Implementation',
                        content: '',
                        locationProvider: null,
                        reactElement: (
                            <FileLocations
                                className="panel__tabs-content"
                                query={this.queryImplementation}
                                inputRepo={subject.repoPath}
                                inputRevision={subject.rev}
                                // tslint:disable-next-line:jsx-no-lambda
                                onSelect={() => this.onSelectLocation('impl')}
                                icon={RepositoryIcon}
                                pluralNoun="implementations"
                                isLightTheme={this.props.isLightTheme}
                                fetchHighlightedFileLines={fetchHighlightedFileLines}
                            />
                        ),
                    }))
                )
            )
        )

        // File history view.
        this.subscriptions.add(
            this.props.extensionsController.services.views.registerProvider(
                { id: 'history', container: ContributableViewContainer.Panel },
                subjectChanges.pipe(
                    map((subject: PanelSubject) => ({
                        title: 'History',
                        content: '',
                        locationProvider: null,
                        reactElement: (
                            <RepoRevSidebarCommits
                                key="commits"
                                repoName={subject.repoPath}
                                repoID={this.props.repoID}
                                rev={subject.rev}
                                filePath={subject.filePath}
                                history={this.props.history}
                                location={this.props.location}
                            />
                        ),
                    }))
                )
            )
        )

        // Code discussions view.
        this.subscriptions.add(
            this.props.extensionsController.services.views.registerProvider(
                { id: 'discussions', container: ContributableViewContainer.Panel },
                subjectChanges.pipe(
                    map(
                        (subject: PanelSubject) =>
                            isDiscussionsEnabled(this.props.settingsCascade)
                                ? {
                                      title: 'Discussions',
                                      content: '',
                                      locationProvider: null,
                                      reactElement: (
                                          <DiscussionsTree
                                              repoID={this.props.repoID}
                                              repoPath={subject.repoPath}
                                              commitID={subject.commitID}
                                              rev={subject.rev}
                                              filePath={subject.filePath}
                                              history={this.props.history}
                                              location={this.props.location}
                                              authenticatedUser={this.props.authenticatedUser}
                                          />
                                      ),
                                  }
                                : null
                    )
                )
            )
        )

        // Update references when subject changes after the initial mount.
        this.subscriptions.add(subjectChanges.pipe(skip(1)).subscribe(() => this.locationsUpdates.next()))
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return null
    }

    private onSelectLocation = (tab: BlobPanelTabID): void => eventLogger.log('BlobPanelLocationSelected', { tab })

    private queryDefinition = (): Observable<{ loading: boolean; locations: Location[] }> =>
        getDefinition(this.props as LSPTextDocumentPositionParams, this.props).pipe(
            map(locations => ({ loading: false, locations: locations ? castArray(locations) : [] }))
        )

    private queryReferencesLocal = (): Observable<{ loading: boolean; locations: Location[] }> =>
        getReferences({ ...(this.props as LSPTextDocumentPositionParams), includeDeclaration: false }, this.props).pipe(
            map(c => ({ loading: false, locations: c }))
        )

    private queryReferencesExternal = (): Observable<{ loading: boolean; locations: Location[] }> =>
        concat(
            fetchExternalReferences(this.props as LSPTextDocumentPositionParams).pipe(
                map(c => ({ loading: true, locations: c }))
            ),
            [{ loading: false, locations: [] }]
        ).pipe(
            bufferTime(500), // reduce UI jitter
            scan<{ loading: boolean; locations: Location[] }[], { loading: boolean; locations: Location[] }>(
                (cur, locs) => ({
                    loading: cur.loading && locs.every(({ loading }) => loading),
                    locations: cur.locations.concat(...locs.map(({ locations }) => locations)),
                }),
                { loading: true, locations: [] }
            )
        )

    private queryImplementation = (): Observable<{ loading: boolean; locations: Location[] }> =>
        getImplementations(this.props as LSPTextDocumentPositionParams, this.props).pipe(
            map(c => ({ loading: false, locations: c }))
        )
}

function renderHoverContents(contents: HoverMerged['contents'][0]): React.ReactFragment {
    const value = typeof contents === 'string' ? contents : contents.value
    const language = typeof contents === 'string' ? 'markdown' : 'kind' in contents ? contents.kind : contents.language
    try {
        if (language === 'markdown') {
            return (
                <div
                    dangerouslySetInnerHTML={{
                        __html: marked(value, { gfm: true, breaks: true, sanitize: true }),
                    }}
                />
            )
        }
        return <code className="hljs" dangerouslySetInnerHTML={{ __html: highlight(language, value).value }} />
    } catch (e) {
        // Ignore rendering or syntax highlighting errors; plain text will be rendered.
    }
    return value
}
