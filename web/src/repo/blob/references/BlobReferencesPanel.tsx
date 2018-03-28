import CloseIcon from '@sourcegraph/icons/lib/Close'
import GlobeIcon from '@sourcegraph/icons/lib/Globe'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import * as H from 'history'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { bufferTime } from 'rxjs/operators/bufferTime'
import { concat } from 'rxjs/operators/concat'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { scan } from 'rxjs/operators/scan'
import { skip } from 'rxjs/operators/skip'
import { startWith } from 'rxjs/operators/startWith'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { Location } from 'vscode-languageserver-types'
import { fetchReferences } from '../../../backend/lsp'
import { Spacer, Tab, TabBorderClassName, TabsWithURLViewStatePersistence } from '../../../components/Tabs'
import { eventLogger } from '../../../tracking/eventLogger'
import { parseHash } from '../../../util/url'
import { AbsoluteRepoFilePosition } from '../../index'
import { FileLocations } from '../panel/FileLocations'
import { fetchExternalReferences } from './backend'

interface Props extends AbsoluteRepoFilePosition {
    location: H.Location
    history: H.History
    isLightTheme: boolean
}

/** The references' subject (what the references refer to). */
interface ReferencesStateSubject {
    repoPath: string
    commitID: string
    filePath: string
    line: number
    character: number
}

type ReferencesTabID = 'references' | 'references:external'

function referencesStateSubject(props: Props): ReferencesStateSubject {
    const parsedHash = parseHash(props.location.hash)
    return {
        repoPath: props.repoPath,
        commitID: props.commitID,
        filePath: props.filePath,
        line: parsedHash.line || 1,
        character: parsedHash.character || 1,
    }
}

function referencesStateSubjectIsEqual(
    a: ReferencesStateSubject,
    b: ReferencesStateSubject & { line?: number; character?: number }
): boolean {
    return (
        a &&
        b &&
        a.repoPath === b.repoPath &&
        a.commitID === b.commitID &&
        a.filePath === b.filePath &&
        a.line === b.line &&
        a.character === b.character
    )
}

interface State {
    referencesCount?: number
    referencesExternalCount?: number
}

/**
 * A panel on the blob page that displays references.
 */
export class BlobReferencesPanel extends React.PureComponent<Props, State> {
    public state: State = {}

    private componentUpdates = new Subject<Props>()
    private locationsUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const componentUpdates = this.componentUpdates.pipe(startWith(this.props))

        // Update references when subject changes.
        this.subscriptions.add(
            componentUpdates
                .pipe(
                    distinctUntilChanged<Props>((a, b) =>
                        referencesStateSubjectIsEqual(referencesStateSubject(a), referencesStateSubject(b))
                    ),
                    skip(1),
                    tap(() => this.setState({ referencesCount: undefined, referencesExternalCount: undefined }))
                )
                .subscribe(() => this.locationsUpdates.next())
        )
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const tabs: Tab<ReferencesTabID>[] = [
            {
                id: 'references',
                label: (
                    <>
                        This repository
                        {this.state.referencesCount !== undefined && (
                            <span className="badge badge-pill badge-secondary ml-1">{this.state.referencesCount}</span>
                        )}
                    </>
                ),
            },
            {
                id: 'references:external',
                label: (
                    <>
                        Other repositories
                        {this.state.referencesExternalCount !== undefined && (
                            <span className="badge badge-pill badge-secondary ml-1">
                                {this.state.referencesExternalCount}
                            </span>
                        )}
                    </>
                ),
            },
        ]

        return (
            <TabsWithURLViewStatePersistence
                tabs={tabs}
                tabBarEndFragment={
                    <>
                        <Spacer />
                        <button
                            onClick={this.onDismiss}
                            className={`btn btn-icon tab-bar__close-button ${TabBorderClassName}`}
                            data-tooltip="Close"
                        >
                            <CloseIcon />
                        </button>
                    </>
                }
                className="blob-references-panel"
                tabClassName="tab-bar__tab--h5like"
                onSelectTab={this.onSelectTab}
                location={this.props.location}
            >
                <FileLocations
                    key="references"
                    className="blob-references-panel__content"
                    query={this.queryReferencesLocal}
                    updates={this.locationsUpdates}
                    inputRevision={this.props.rev}
                    icon={RepoIcon}
                    onSelect={this.logLocalSelection}
                    pluralNoun="local references"
                    isLightTheme={this.props.isLightTheme}
                />
                <FileLocations
                    key="references:external"
                    className="blob-references-panel__content"
                    query={this.queryReferencesExternal}
                    updates={this.locationsUpdates}
                    icon={GlobeIcon}
                    onSelect={this.logExternalSelection}
                    pluralNoun="external references"
                    isLightTheme={this.props.isLightTheme}
                />
            </TabsWithURLViewStatePersistence>
        )
    }

    private onDismiss = (): void => {
        this.props.history.push(TabsWithURLViewStatePersistence.urlForTabID(this.props.location, null))
    }

    private onSelectTab = (tab: string): void => {
        if (tab === 'references') {
            eventLogger.log('ShowLocalRefsButtonClicked')
        } else if (tab === 'references:external') {
            eventLogger.log('ShowExternalRefsButtonClicked')
        }
    }

    private logLocalSelection = () => eventLogger.log('GoToLocalRefClicked')
    private logExternalSelection = () => eventLogger.log('GoToExternalRefClicked')

    private queryReferencesLocal = (): Observable<{ loading: boolean; locations: Location[] }> =>
        fetchReferences(this.props).pipe(
            map(c => ({ loading: false, locations: c })),
            tap(({ locations }) => this.setState({ referencesCount: locations.length }))
        )

    private queryReferencesExternal = (): Observable<{ loading: boolean; locations: Location[] }> =>
        fetchExternalReferences(this.props).pipe(
            map(c => ({ loading: true, locations: c })),
            concat([{ loading: false, locations: [] }]),
            bufferTime(500), // reduce UI jitter
            scan<{ loading: boolean; locations: Location[] }[], { loading: boolean; locations: Location[] }>(
                (cur, locs) => ({
                    loading: cur.loading && locs.every(({ loading }) => loading),
                    locations: cur.locations.concat(...locs.map(({ locations }) => locations)),
                }),
                { loading: true, locations: [] }
            ),
            tap(({ loading, locations }) => {
                if (!loading || locations.length > 0) {
                    this.setState({ referencesExternalCount: locations.length })
                }
            })
        )
}
