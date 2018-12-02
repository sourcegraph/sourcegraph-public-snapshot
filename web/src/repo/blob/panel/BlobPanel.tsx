import { highlight } from 'highlight.js/lib/highlight'
import * as H from 'history'
import { castArray, isEqual } from 'lodash'
import marked from 'marked'
import * as React from 'react'
import { merge, Observable, of, Subject, Subscription } from 'rxjs'
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
import { Location } from '../../../../../shared/src/api/protocol/plainTypes'
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

export type BlobPanelTabID = 'info' | 'def' | 'references' | 'references:external' | 'discussions' | 'impl' | 'history'

/** The subject (what the contextual information refers to). */
interface PanelSubject extends AbsoluteRepoFile, ModeSpec, Partial<PositionSpec> {
    repoID: string
}

function toSubject(props: Props): PanelSubject {
    const parsedHash = parseHash(props.location.hash)
    return {
        repoPath: props.repoPath,
        repoID: props.repoID,
        commitID: props.commitID,
        rev: props.rev,
        filePath: props.filePath,
        mode: props.mode,
        position:
            parsedHash.line !== undefined ? { line: parsedHash.line, character: parsedHash.character || 0 } : undefined,
    }
}

const LOADING: 'loading' = 'loading'

/**
 * A panel on the blob page that displays contextual information.
 */
export class BlobPanel extends React.PureComponent<Props> {
    private componentUpdates = new Subject<Props>()
    private locationsUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    public constructor(props: Props) {
        super(props)

        const componentUpdates = this.componentUpdates.pipe(startWith(this.props))

        // Changes to the subject, including upon the initial mount.
        const subjectChanges = componentUpdates.pipe(
            map(toSubject),
            distinctUntilChanged((a, b) => isEqual(a, b))
        )

        this.subscriptions.add(
            this.props.extensionsController.services.views.registerProviders([
                {
                    // Info (hover) panel view.
                    registrationOptions: { id: 'Info', container: ContributableViewContainer.Panel },
                    provider: subjectChanges.pipe(
                        switchMap((subject: PanelSubject) => {
                            if (!subject.position) {
                                return [null]
                            }
                            const result = getHover(subject as LSPTextDocumentPositionParams, {
                                extensionsController: this.props.extensionsController,
                            }).pipe(catchError(error => [asError(error) as ErrorLike]))
                            return merge(
                                result,
                                // Delay loading spinner to reduce jitter.
                                of(LOADING).pipe(
                                    delay(150),
                                    takeUntil(result)
                                )
                            ).pipe(
                                // Clear old data immediately.
                                startWith(null)
                            )
                        }),
                        map((hoverOrError: null | HoverMerged | ErrorLike | typeof LOADING) => {
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
                    ),
                },

                {
                    // Definition panel view.
                    registrationOptions: { id: 'def', container: ContributableViewContainer.Panel },
                    provider: subjectChanges.pipe(
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
                    ),
                },

                {
                    // References panel view.
                    registrationOptions: { id: 'references', container: ContributableViewContainer.Panel },
                    provider: subjectChanges.pipe(
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
                    ),
                },

                {
                    // External references panel view.
                    registrationOptions: { id: 'references:external', container: ContributableViewContainer.Panel },
                    provider: subjectChanges.pipe(
                        map(() => ({
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
                    ),
                },

                {
                    // Implementations panel view.
                    registrationOptions: { id: 'impl', container: ContributableViewContainer.Panel },
                    provider: subjectChanges.pipe(
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
                    ),
                },

                {
                    // File history view.
                    registrationOptions: { id: 'history', container: ContributableViewContainer.Panel },
                    provider: subjectChanges.pipe(
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
                    ),
                },

                {
                    // Code discussions view.
                    registrationOptions: { id: 'discussions', container: ContributableViewContainer.Panel },
                    provider: subjectChanges.pipe(
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
                    ),
                },
            ])
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

    private queryDefinition = (): Observable<Location[]> =>
        getDefinition(this.props as LSPTextDocumentPositionParams, this.props).pipe(
            map(locations => castArray(locations || []))
        )

    private queryReferencesLocal = (): Observable<Location[]> =>
        getReferences({ ...(this.props as LSPTextDocumentPositionParams), includeDeclaration: false }, this.props)

    private queryReferencesExternal = (): Observable<Location[]> =>
        fetchExternalReferences(this.props as LSPTextDocumentPositionParams).pipe(
            bufferTime(500), // reduce UI jitter
            scan<Location[][], Location[]>((cur, locs) => cur.concat(...locs.map(locations => locations)), [])
        )

    private queryImplementation = (): Observable<Location[]> =>
        getImplementations(this.props as LSPTextDocumentPositionParams, this.props)
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
