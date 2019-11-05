import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { from, Observable, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, startWith, switchMap, tap } from 'rxjs/operators'
import { getActiveCodeEditorPosition } from '../../../../../shared/src/api/client/services/editorService'
import { TextDocumentLocationProviderRegistry } from '../../../../../shared/src/api/client/services/location'
import { Entry } from '../../../../../shared/src/api/client/services/registry'
import {
    PanelViewWithComponent,
    ProvideViewSignature,
    ViewProviderRegistrationOptions,
} from '../../../../../shared/src/api/client/services/view'
import { ContributableViewContainer, TextDocumentPositionParams } from '../../../../../shared/src/api/protocol'
import { ActivationProps } from '../../../../../shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { AbsoluteRepoFile, ModeSpec, parseHash, PositionSpec } from '../../../../../shared/src/util/url'
import { isDiscussionsEnabled } from '../../../discussions'
import { RepoHeaderContributionsLifecycleProps } from '../../RepoHeader'
import { RepoRevSidebarCommits } from '../../RepoRevSidebarCommits'
import { DiscussionsTree } from '../discussions/DiscussionsTree'
import { ThemeProps } from '../../../../../shared/src/theme'
interface Props
    extends AbsoluteRepoFile,
        Partial<PositionSpec>,
        ModeSpec,
        RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        PlatformContextProps,
        ExtensionsControllerProps,
        ThemeProps,
        ActivationProps {
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

    /**
     * Include the full URI fragment here because it represents the state of panels, and we want
     * panels to be re-rendered when this state changes.
     */
    hash: string
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
        hash: props.location.hash,
    }
}

/**
 * A null-rendered component that registers panel views for the blob.
 */
export class BlobPanel extends React.PureComponent<Props> {
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        const componentUpdates = this.componentUpdates.pipe(startWith(this.props))

        // Changes to the subject, including upon the initial mount.
        const subjectChanges = componentUpdates.pipe(
            map(toSubject),
            distinctUntilChanged((a, b) => isEqual(a, b))
        )

        const entryForViewProviderRegistration = <P extends TextDocumentPositionParams>(
            id: string,
            title: string,
            priority: number,
            registry: TextDocumentLocationProviderRegistry<P>,
            extraParams?: Pick<P, Exclude<keyof P, keyof TextDocumentPositionParams>>
        ): Entry<ViewProviderRegistrationOptions, ProvideViewSignature> => ({
            registrationOptions: { id, container: ContributableViewContainer.Panel },
            provider: from(this.props.extensionsController.services.editor.activeEditorUpdates).pipe(
                map(activeEditor =>
                    activeEditor
                        ? {
                              ...activeEditor,
                              model: this.props.extensionsController.services.model.getPartialModel(
                                  activeEditor.resource
                              ),
                          }
                        : undefined
                ),
                switchMap(activeEditor =>
                    registry.hasProvidersForActiveTextDocument(activeEditor).pipe(
                        map(hasProviders => {
                            if (!hasProviders) {
                                return null
                            }
                            const params: TextDocumentPositionParams | null = getActiveCodeEditorPosition(activeEditor)
                            if (!params) {
                                return null
                            }
                            return {
                                title,
                                content: '',
                                priority,

                                // This disable directive is necessary because TypeScript is not yet smart
                                // enough to know that (typeof params & typeof extraParams) is P.
                                //
                                // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
                                locationProvider: registry.getLocations({ ...params, ...extraParams } as P).pipe(
                                    map(locationsObservable =>
                                        locationsObservable.pipe(
                                            tap(locations => {
                                                if (
                                                    this.props.activation &&
                                                    id === 'references' &&
                                                    locations &&
                                                    locations.length > 0
                                                ) {
                                                    this.props.activation.update({ FoundReferences: true })
                                                }
                                            })
                                        )
                                    )
                                ),
                            }
                        })
                    )
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
                                                  compact={true}
                                                  extensionsController={this.props.extensionsController}
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
