import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { from, Observable, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, skip, startWith } from 'rxjs/operators'
import { TextDocumentLocationProviderRegistry } from '../../../../../shared/src/api/client/services/location'
import { Entry } from '../../../../../shared/src/api/client/services/registry'
import {
    PanelViewWithComponent,
    ProvideViewSignature,
    ViewProviderRegistrationOptions,
} from '../../../../../shared/src/api/client/services/view'
import { ContributableViewContainer, TextDocumentPositionParams } from '../../../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { AbsoluteRepoFile, ModeSpec, parseHash, PositionSpec } from '../../../../../shared/src/util/url'
import { isDiscussionsEnabled } from '../../../discussions'
import { ThemeProps } from '../../../theme'
import { RepoHeaderContributionsLifecycleProps } from '../../RepoHeader'
import { RepoRevSidebarCommits } from '../../RepoRevSidebarCommits'
import { DiscussionsTree } from '../discussions/DiscussionsTree'
interface Props
    extends AbsoluteRepoFile,
        Partial<PositionSpec>,
        ModeSpec,
        RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        PlatformContextProps,
        ExtensionsControllerProps,
        ThemeProps {
    location: H.Location
    history: H.History
    repoID: GQL.ID
    repoName: string
    commitID: string
    authenticatedUser: GQL.IUser | null
}

export type BlobPanelTabID = 'info' | 'def' | 'references' | 'discussions' | 'impl' | 'typedef' | 'history'

/** The subject (what the contextual information refers to). */
interface PanelSubject extends AbsoluteRepoFile, ModeSpec, Partial<PositionSpec> {
    repoID: string
}

function toSubject(props: Props): PanelSubject {
    const parsedHash = parseHash(props.location.hash)
    return {
        repoName: props.repoName,
        repoID: props.repoID,
        commitID: props.commitID,
        rev: props.rev,
        filePath: props.filePath,
        mode: props.mode,
        position:
            parsedHash.line !== undefined ? { line: parsedHash.line, character: parsedHash.character || 0 } : undefined,
    }
}

/**
 * A null-rendered component that registers panel views for the blob.
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
            priority: number,
            registry: TextDocumentLocationProviderRegistry<P>,
            extraParams?: Pick<P, Exclude<keyof P, keyof TextDocumentPositionParams>>
        ) => Entry<ViewProviderRegistrationOptions, ProvideViewSignature> = (
            id,
            title,
            priority,
            registry,
            extraParams
        ) => ({
            registrationOptions: { id, container: ContributableViewContainer.Panel },
            provider: registry
                .getLocationsAndProviders(from(this.props.extensionsController.services.model.model), extraParams)
                .pipe(
                    map(({ locations, hasProviders }) =>
                        hasProviders && locations
                            ? {
                                  title,
                                  content: '',
                                  priority,
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
                        190,
                        this.props.extensionsController.services.textDocumentDefinition
                    ),
                    entryForViewProviderRegistration(
                        'references',
                        'References',
                        180,
                        this.props.extensionsController.services.textDocumentReferences,
                        {
                            context: { includeDeclaration: false },
                        }
                    ),
                    entryForViewProviderRegistration(
                        'impl',
                        'Implementation',
                        160,
                        this.props.extensionsController.services.textDocumentImplementation
                    ),
                    entryForViewProviderRegistration(
                        'typedef',
                        'Type definition',
                        150,
                        this.props.extensionsController.services.textDocumentTypeDefinition
                    ),

                    {
                        // File history view.
                        registrationOptions: { id: 'history', container: ContributableViewContainer.Panel },
                        provider: subjectChanges.pipe(
                            map((subject: PanelSubject) => ({
                                title: 'History',
                                content: '',
                                priority: 150,
                                locationProvider: null,
                                reactElement: (
                                    <RepoRevSidebarCommits
                                        key="commits"
                                        repoName={subject.repoName}
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
                            map((subject: PanelSubject) =>
                                isDiscussionsEnabled(this.props.settingsCascade)
                                    ? {
                                          title: 'Discussions',
                                          content: '',
                                          priority: 140,
                                          locationProvider: null,
                                          reactElement: (
                                              <DiscussionsTree
                                                  repoID={this.props.repoID}
                                                  repoName={subject.repoName}
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
}
