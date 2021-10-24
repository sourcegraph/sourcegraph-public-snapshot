import classNames from 'classnames'
import * as H from 'history'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import * as React from 'react'
import { Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, endWith, map, startWith, switchMap, tap } from 'rxjs/operators'

import { MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { Location } from '@sourcegraph/extension-api-types'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { Resizable } from '@sourcegraph/shared/src/components/Resizable'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { parseRepoURI } from '@sourcegraph/shared/src/util/url'

import { FileLocations, FileLocationsError, FileLocationsNotFound } from './FileLocations'
import styles from './HierarchicalLocationsView.module.scss'
import { HierarchicalLocationsViewButton } from './HierarchicalLocationsViewButton'
import { groupLocations } from './locations'

/** The maximum number of results we'll receive from a provider before we truncate and display a banner. */
const MAXIMUM_LOCATION_RESULTS = 500

export interface HierarchicalLocationsViewProps
    extends SettingsCascadeProps,
        TelemetryProps,
        ExtensionsControllerProps<'extHostAPI'> {
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

    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}

interface State {
    /**
     * Locations (inside files identified by LSP-style git:// URIs) to display,
     * loading, or an error if they failed to load.
     *
     * Locations may be truncated if the result set is too large.
     */
    locationsOrError: MaybeLoadingResult<{ locations: Location[]; isTruncated: boolean } | ErrorLike>

    selectedGroups?: string[]
}

interface LocationGroup {
    name: string
    defaultSize: number
    key: (location: Location) => string | undefined
}

/**
 * Displays a multi-column view to drill down (by repository, file, etc.) to a list of locations in files.
 */
export class HierarchicalLocationsView extends React.PureComponent<HierarchicalLocationsViewProps, State> {
    public state: State = { locationsOrError: { isLoading: true, result: { locations: [], isTruncated: false } } }

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
                            // Truncate the result set if it is too large,
                            // to avoid crashing the UI. A banner will be displayed to the user
                            // when this is the case.
                            map(({ isLoading, result: locations }) => {
                                const isTruncated = locations.length > MAXIMUM_LOCATION_RESULTS
                                return {
                                    isLoading,
                                    result: {
                                        locations: isTruncated
                                            ? locations.slice(0, MAXIMUM_LOCATION_RESULTS)
                                            : locations,
                                        isTruncated,
                                    },
                                }
                            }),
                            catchError((error): [State['locationsOrError']] => [
                                { isLoading: false, result: asError(error) },
                            ]),
                            startWith({
                                result: { locations: [], isTruncated: false },
                                isLoading: true,
                            }),
                            tap(({ result }) => {
                                const hasResults = !isErrorLike(result) && result.locations.length > 0
                                this.props.extensionsController.extHostAPI
                                    .then(extensionHostAPI =>
                                        extensionHostAPI.updateContext({
                                            'panel.locations.hasResults': hasResults,
                                        })
                                    )
                                    .catch(() => {
                                        // noop
                                    })
                            }),
                            endWith({ isLoading: false })
                        )
                    )
                )
                .subscribe(locationsOrError =>
                    this.setState(previous => ({
                        locationsOrError: {
                            ...previous.locationsOrError,
                            ...locationsOrError,
                        },
                    }))
                )
        )

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
        if (this.state.locationsOrError.isLoading && this.state.locationsOrError.result.locations.length === 0) {
            return <LoadingSpinner className="icon-inline m-1 test-loading-spinner" />
        }
        if (this.state.locationsOrError.result.locations.length === 0) {
            return <FileLocationsNotFound />
        }

        const GROUPS: LocationGroup[] = [
            {
                name: 'repo',
                defaultSize: 175,
                key: location => parseRepoURI(location.uri).repoName,
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
                key: location => parseRepoURI(location.uri).filePath,
            })
        }

        const { groups, selectedGroups, visibleLocations } = groupLocations(
            this.state.locationsOrError.result.locations,
            this.state.selectedGroups || null,
            GROUPS.map(({ key }) => key),
            { uri: this.props.defaultGroup }
        )

        const groupsToDisplay = GROUPS.map(({ name, key, defaultSize }, index) => {
            const group = { name, key, defaultSize }
            if (!groups[index]) {
                // No groups exist at this level. Don't display anything.
                return null
            }
            if (groups[index].length > 1) {
                // Always display when there is more than 1 group.
                return group
            }
            if (groups[index].length === 1) {
                if (selectedGroups[index] !== groups[index][0].key) {
                    // When the only group is not the currently selected group, show it. This occurs when the
                    // references list changes after the user made an initial selection. The group must be shown so
                    // that the user can update their selection to the only available group; otherwise they would
                    // be stuck viewing the (zero) results from the previously selected group that no longer
                    // exists.
                    return group
                }
                if (key({ uri: this.props.defaultGroup }) !== selectedGroups[index]) {
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
            <div>
                {this.state.locationsOrError.result.isTruncated && (
                    <div className="alert alert-warning py-1 px-3 m-2 text-nowrap text-center">
                        <small>
                            <strong>Large result set</strong> - only showing the first {MAXIMUM_LOCATION_RESULTS}{' '}
                            results.
                        </small>
                    </div>
                )}
                <div
                    className={classNames(styles.referencesContainer, this.props.className)}
                    data-testid="hierarchical-locations-view"
                >
                    <div className="d-flex">
                        {selectedGroups &&
                            groupsToDisplay.map(
                                (group, index) =>
                                    group && (
                                        <Resizable
                                            key={index}
                                            className={styles.resizableGroup}
                                            handleClassName={styles.resizableHandle}
                                            handlePosition="right"
                                            storageKey={`hierarchical-locations-view-resizable:${group.name}`}
                                            defaultSize={group.defaultSize}
                                            element={
                                                <div
                                                    data-testid="hierarchical-locations-view-list"
                                                    className={classNames('list-group', styles.groupList)}
                                                >
                                                    {groups[index].map((group, innerIndex) => (
                                                        <HierarchicalLocationsViewButton
                                                            key={innerIndex}
                                                            groupKey={group.key}
                                                            groupCount={group.count}
                                                            isActive={selectedGroups[index] === group.key}
                                                            onClick={event =>
                                                                this.onSelectTree(
                                                                    event,
                                                                    selectedGroups,
                                                                    index,
                                                                    group.key
                                                                )
                                                            }
                                                        />
                                                    ))}
                                                    {this.state.locationsOrError.isLoading && (
                                                        <LoadingSpinner className="icon-inline m-2 flex-shrink-0 test-loading-spinner" />
                                                    )}
                                                </div>
                                            }
                                        />
                                    )
                            )}
                    </div>
                    <FileLocations
                        className={styles.fileLocations}
                        location={this.props.location}
                        telemetryService={this.props.telemetryService}
                        locations={of(visibleLocations)}
                        onSelect={this.props.onSelectLocation}
                        icon={FileDocumentIcon}
                        fetchHighlightedFileLineRanges={this.props.fetchHighlightedFileLineRanges}
                        settingsCascade={this.props.settingsCascade}
                        parentContainerIsEmpty={this.state.locationsOrError.result.locations.length === 0}
                    />
                </div>
            </div>
        )
    }

    private onSelectTree = (
        event: React.MouseEvent<HTMLElement>,
        selectedGroups: string[],
        index: number,
        group: string
    ): void => {
        event.preventDefault()
        this.setState({ selectedGroups: selectedGroups.slice(0, index).concat(group) })
        if (this.props.onSelectTree) {
            this.props.onSelectTree()
        }
    }
}
