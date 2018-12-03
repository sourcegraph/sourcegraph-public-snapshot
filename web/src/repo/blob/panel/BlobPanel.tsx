import { highlight } from 'highlight.js/lib/highlight'
import * as H from 'history'
import { isEqual } from 'lodash'
import marked from 'marked'
import * as React from 'react'
import { from, merge, Observable, of, Subject, Subscription } from 'rxjs'
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
import { TextDocumentLocationProviderRegistry } from '../../../../../shared/src/api/client/services/location'
import { Entry } from '../../../../../shared/src/api/client/services/registry'
import {
    PanelViewWithComponent,
    ProvideViewSignature,
    ViewProviderRegistrationOptions,
} from '../../../../../shared/src/api/client/services/view'
import { ContributableViewContainer, TextDocumentPositionParams } from '../../../../../shared/src/api/protocol'
import { Location } from '../../../../../shared/src/api/protocol/plainTypes'
import { RepositoryIcon } from '../../../../../shared/src/components/icons' // TODO: Switch to mdi icon
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { HierarchicalLocationsView } from '../../../../../shared/src/panel/views/HierarchicalLocationsView'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { AbsoluteRepoFile, parseHash, PositionSpec } from '../../../../../shared/src/util/url'
import { getHover, HoverMerged, ModeSpec } from '../../../backend/features'
import { isEmptyHover, LSPTextDocumentPositionParams } from '../../../backend/lsp'
import { isDiscussionsEnabled } from '../../../discussions'
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

const showExternalReferences = localStorage.getItem('hideExternalReferencesPanel') === null

export type BlobPanelTabID =
    | 'info'
    | 'def'
    | 'references'
    | 'references:external'
    | 'discussions'
    | 'impl'
    | 'typedef'
    | 'history'

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

        const entryForViewProviderRegistration: <P extends TextDocumentPositionParams>(
            id: string,
            title: string,
            registry: TextDocumentLocationProviderRegistry<P>,
            extraParams?: Pick<P, Exclude<keyof P, keyof TextDocumentPositionParams>>
        ) => Entry<ViewProviderRegistrationOptions, ProvideViewSignature> = (id, title, registry, extraParams) => ({
            registrationOptions: { id, container: ContributableViewContainer.Panel },
            provider: registry
                .getLocationsAndProviders(from(this.props.extensionsController.services.model.model), extraParams)
                .pipe(
                    map(
                        ({ locations, hasProviders }) =>
                            hasProviders && locations
                                ? {
                                      title,
                                      content: '',
                                      locationProvider: locations,
                                  }
                                : null
                    )
                ),
        })

        this.subscriptions.add(
            this.props.extensionsController.services.views.registerProviders(
                [
                    entryForViewProviderRegistration(
                        'def',
                        'Definition',
                        this.props.extensionsController.services.textDocumentDefinition
                    ),
                    entryForViewProviderRegistration(
                        'references',
                        'References',
                        this.props.extensionsController.services.textDocumentReferences,
                        {
                            context: { includeDeclaration: false },
                        }
                    ),
                    entryForViewProviderRegistration(
                        'impl',
                        'Implementation',
                        this.props.extensionsController.services.textDocumentImplementation
                    ),
                    entryForViewProviderRegistration(
                        'typedef',
                        'Type definition',
                        this.props.extensionsController.services.textDocumentTypeDefinition
                    ),
                    {
                        // Info (hover) panel view.
                        registrationOptions: { id: 'Info', container: ContributableViewContainer.Panel },
                        provider: subjectChanges.pipe(
                            switchMap((subject: PanelSubject) => {
                                if (!subject.position || subject.position.character === 0) {
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

                    showExternalReferences
                        ? {
                              // External references panel view.
                              registrationOptions: {
                                  id: 'references:external',
                                  container: ContributableViewContainer.Panel,
                              },
                              provider: subjectChanges.pipe(
                                  map(() => ({
                                      title: 'External references',
                                      content: '',
                                      locationProvider: null,
                                      reactElement: (
                                          <HierarchicalLocationsView
                                              className="panel__tabs-content"
                                              locations={this.queryReferencesExternal()}
                                              icon={RepositoryIcon}
                                              isLightTheme={this.props.isLightTheme}
                                              fetchHighlightedFileLines={fetchHighlightedFileLines}
                                          />
                                      ),
                                  }))
                              ),
                          }
                        : null,

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
                ].filter(
                    (v): v is Entry<ViewProviderRegistrationOptions, Observable<PanelViewWithComponent | null>> => !!v
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

    private queryReferencesExternal = (): Observable<Location[]> =>
        fetchExternalReferences(this.props as LSPTextDocumentPositionParams).pipe(
            bufferTime(500), // reduce UI jitter
            scan<Location[][], Location[]>((cur, locs) => cur.concat(...locs.map(locations => locations)), [])
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
