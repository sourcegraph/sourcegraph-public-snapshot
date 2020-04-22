import { Location } from '@sourcegraph/extension-api-types'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import H from 'history'
import * as React from 'react'
import { Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, endWith, map, startWith, switchMap, tap } from 'rxjs/operators'
import { FetchFileCtx } from '../../components/CodeExcerpt'
import { RepoLink } from '../../components/RepoLink'
import { Resizable } from '../../components/Resizable'
import { ExtensionsControllerProps } from '../../extensions/controller'
import { SettingsCascadeProps } from '../../settings/settings'
import { asError, ErrorLike, isErrorLike } from '../../util/errors'
import { parseRepoURI } from '../../util/url'
import { registerPanelToolbarContributions } from './contributions'
import { FileLocations, FileLocationsError, FileLocationsNotFound } from './FileLocations'
import { groupLocations } from './locations'
import { MaybeLoadingResult } from '@sourcegraph/codeintellify'

export interface HierarchicalLocationsViewProps extends ExtensionsControllerProps<'services'>, SettingsCascadeProps {
    location: H.Location
    /**
     * The observable that emits the locations.
     */
    locations: Observable<MaybeLoadingResult<Location[]>>

    /**
     * In the grouping (i.e., by repository and, optionally, then by file), this is the URI of the first group.
     * Usually this is set to the URI to the root of the repository that is currently being viewed to ensure that
     * it is listed first.
     */
    defaultGroup: string

    /** Called when an item in the tree is selected. */
    onSelectTree?: () => void

    /** Called when a location is selected. */
    onSelectLocation?: () => void

    className?: string

    isLightTheme: boolean

    fetchHighlightedFileLines: (ctx: FetchFileCtx, force?: boolean) => Observable<string[]>
}

interface State {
    /**
     * Locations (inside files identified by LSP-style git:// URIs) to display, loading, or an error if they failed
     * to load.
     */
    locationsOrError: MaybeLoadingResult<Location[] | ErrorLike>

    selectedGroups?: string[]
}

/**
 * Displays a multi-column view to drill down (by repository, file, etc.) to a list of locations in files.
 */
export class HierarchicalLocationsView extends React.PureComponent<HierarchicalLocationsViewProps, State> {
    public state: State = { locationsOrError: { isLoading: true, result: [] } }

    private componentUpdates = new Subject<HierarchicalLocationsViewProps>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const locationProvidersChanges = this.componentUpdates.pipe(
            map(({ locations }) => locations),
            distinctUntilChanged()
        )

        this.subscriptions.add(
            locationProvidersChanges
                .pipe(
                    switchMap(locationProviderResults =>
                        locationProviderResults.pipe(
                            catchError((error): [MaybeLoadingResult<Location[] | ErrorLike>] => [
                                { isLoading: false, result: asError(error) },
                            ]),
                            startWith({ result: [], isLoading: true }),
                            tap(({ result }) => {
                                const hasResults = !isErrorLike(result) && result.length > 0
                                this.props.extensionsController.services.context.updateContext({
                                    'panel.locations.hasResults': hasResults,
                                })
                            }),
                            endWith({ isLoading: false })
                        )
                    )
                )
                .subscribe(locationsOrError =>
                    this.setState(previous => ({
                        locationsOrError: { ...previous.locationsOrError, ...locationsOrError },
                    }))
                )
        )

        this.subscriptions.add(registerPanelToolbarContributions(this.props))

        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (isErrorLike(this.state.locationsOrError.result)) {
            return <FileLocationsError error={this.state.locationsOrError.result} />
        }
        if (this.state.locationsOrError.isLoading && this.state.locationsOrError.result.length === 0) {
            return <LoadingSpinner className="icon-inline m-1 e2e-loading-spinner" />
        }
        if (this.state.locationsOrError.result.length === 0) {
            return <FileLocationsNotFound />
        }

        const GROUPS: {
            name: string
            defaultSize: number
            key: (loc: Location) => string | undefined
        }[] = [
            {
                name: 'repo',
                defaultSize: 175,
                key: loc => parseRepoURI(loc.uri).repoName,
            },
        ]
        const groupByFile =
            this.props.settingsCascade.final &&
            !isErrorLike(this.props.settingsCascade.final) &&
            this.props.settingsCascade.final['panel.locations.groupByFile']
        if (groupByFile) {
            GROUPS.push({
                name: 'file',
                defaultSize: 200,
                key: loc => parseRepoURI(loc.uri).filePath,
            })
        }

        const { groups, selectedGroups, visibleLocations } = groupLocations<Location, string>(
            this.state.locationsOrError.result,
            this.state.selectedGroups || null,
            GROUPS.map(({ key }) => key),
            { uri: this.props.defaultGroup }
        )

        const groupsToDisplay = GROUPS.map(({ name, key, defaultSize }, i) => {
            const group = { name, key, defaultSize }
            if (!groups[i]) {
                // No groups exist at this level. Don't display anything.
                return null
            }
            if (groups[i].length > 1) {
                // Always display when there is more than 1 group.
                return group
            }
            if (groups[i].length === 1) {
                if (selectedGroups[i] !== groups[i][0].key) {
                    // When the only group is not the currently selected group, show it. This occurs when the
                    // references list changes after the user made an initial selection. The group must be shown so
                    // that the user can update their selection to the only available group; otherwise they would
                    // be stuck viewing the (zero) results from the previously selected group that no longer
                    // exists.
                    return group
                }
                if (key({ uri: this.props.defaultGroup }) !== selectedGroups[i]) {
                    // When the only group is other than the default group, show it. This is important because it
                    // often indicates that the match comes from another repository. If it isn't shown, the user
                    // would likely assume the match is from the current repository.
                    return group
                }
            }
            if (groupByFile && name === 'file') {
                // Always display the file groups when group-by-file is enabled.
                return group
            }
            return null
        })

        return (
            <div className={`hierarchical-locations-view ${this.props.className || ''}`}>
                {selectedGroups &&
                    groupsToDisplay.map(
                        (g, i) =>
                            g && (
                                <Resizable
                                    key={i}
                                    className="hierarchical-locations-view__resizable"
                                    handlePosition="right"
                                    storageKey={`hierarchical-locations-view-resizable:${g.name}`}
                                    defaultSize={g.defaultSize}
                                    element={
                                        <div className="list-group list-group-flush hierarchical-locations-view__list e2e-hierarchical-locations-view-list">
                                            {groups[i].map((group, j) => (
                                                <span
                                                    key={j}
                                                    className={`list-group-item hierarchical-locations-view__item ${
                                                        selectedGroups[i] === group.key ? 'active' : ''
                                                    }`}
                                                    onClick={e => this.onSelectTree(e, selectedGroups, i, group.key)}
                                                >
                                                    <span
                                                        className="hierarchical-locations-view__item-name"
                                                        title={group.key}
                                                    >
                                                        <span className="hierarchical-locations-view__item-name-text">
                                                            <RepoLink to={null} repoName={group.key} />
                                                        </span>
                                                    </span>
                                                    <span className="badge badge-secondary badge-pill hierarchical-locations-view__item-badge">
                                                        {group.count}
                                                    </span>
                                                </span>
                                            ))}
                                            {this.state.locationsOrError.isLoading && (
                                                <LoadingSpinner className="icon-inline m-2 flex-shrink-0 e2e-loading-spinner" />
                                            )}
                                        </div>
                                    }
                                />
                            )
                    )}
                <FileLocations
                    className="hierarchical-locations-view__content"
                    location={this.props.location}
                    locations={of(visibleLocations)}
                    onSelect={this.props.onSelectLocation}
                    icon={SourceRepositoryIcon}
                    isLightTheme={this.props.isLightTheme}
                    fetchHighlightedFileLines={this.props.fetchHighlightedFileLines}
                    settingsCascade={this.props.settingsCascade}
                />
            </div>
        )
    }

    private onSelectTree = (
        e: React.MouseEvent<HTMLElement>,
        selectedGroups: string[],
        i: number,
        group: string
    ): void => {
        e.preventDefault()
        this.setState({ selectedGroups: selectedGroups.slice(0, i).concat(group) })
        if (this.props.onSelectTree) {
            this.props.onSelectTree()
        }
    }
}
