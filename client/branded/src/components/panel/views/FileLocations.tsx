import * as React from 'react'

import { mdiMapSearch } from '@mdi/js'
import classNames from 'classnames'
import type * as H from 'history'
import { upperFirst } from 'lodash'
import { type Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, startWith, switchMap } from 'rxjs/operators'

import { asError, type ErrorLike, isErrorLike, isDefined, property, logger } from '@sourcegraph/common'
import type { Location } from '@sourcegraph/extension-api-types'
import type { FetchFileParameters } from '@sourcegraph/shared/src/backend/file'
import type { Badged } from '@sourcegraph/shared/src/codeintel/legacy-extensions/api'
import { VirtualList } from '@sourcegraph/shared/src/components/VirtualList'
import type { ContentMatch } from '@sourcegraph/shared/src/search/stream'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { parseRepoURI } from '@sourcegraph/shared/src/util/url'
import { LoadingSpinner, Alert, Icon } from '@sourcegraph/wildcard'

import { FileContentSearchResult } from '../../../search-ui'

import styles from './FileLocations.module.scss'

export const FileLocationsError: React.FunctionComponent<React.PropsWithChildren<{ error: ErrorLike }>> = ({
    error,
}) => (
    <Alert className="m-2" variant="danger">
        Error getting locations: {upperFirst(error.message)}
    </Alert>
)

export const FileLocationsNotFound: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <div className={classNames('m-2', styles.notFound)}>
        <Icon aria-hidden={true} svgPath={mdiMapSearch} /> No locations found
    </div>
)

export const FileLocationsNoGroupSelected: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <div className="m-2">
        <Icon aria-hidden={true} svgPath={mdiMapSearch} /> No locations found in the current repository
    </div>
)

interface Props extends SettingsCascadeProps, TelemetryProps {
    location: H.Location
    /**
     * The observable that emits the locations.
     */
    locations: Observable<Location[] | null>

    /** The icon to use for each location. */
    icon: React.ComponentType<{ className?: string }>

    /** Called when a location is selected. */
    onSelect?: () => void

    className?: string

    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>

    /** Whether or not there are other groups in the parent container with results. */
    parentContainerIsEmpty: boolean
}

const LOADING = 'loading' as const

interface State {
    /**
     * Locations (inside files identified by LSP-style git:// URIs) to display, loading, or an error if they failed
     * to load.
     */
    locationsOrError: typeof LOADING | Location[] | null | ErrorLike

    itemsToShow: number
}

interface OrderedURI {
    uri: string
    repo: string
}

/**
 * Displays a flat list of file excerpts. For a tree view, use FileLocationsTree.
 */
export class FileLocations extends React.PureComponent<Props, State> {
    public state: State = {
        locationsOrError: LOADING,
        itemsToShow: 3,
    }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const locationsChanges = this.componentUpdates.pipe(
            map(({ locations }) => locations),
            distinctUntilChanged()
        )

        this.subscriptions.add(
            locationsChanges
                .pipe(
                    switchMap(query => query.pipe(catchError(error => [asError(error) as ErrorLike]))),
                    startWith(LOADING),
                    map(result => ({ locationsOrError: result }))
                )
                .subscribe(
                    stateUpdate => this.setState(stateUpdate),
                    error => logger.error(error)
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
        if (isErrorLike(this.state.locationsOrError)) {
            return <FileLocationsError error={this.state.locationsOrError} />
        }
        if (this.state.locationsOrError === LOADING) {
            return <LoadingSpinner className="m-1" />
        }
        if (this.state.locationsOrError === null || this.state.locationsOrError.length === 0) {
            return this.props.parentContainerIsEmpty ? <FileLocationsNotFound /> : <FileLocationsNoGroupSelected />
        }

        // Locations by fully qualified URI, like git://github.com/gorilla/mux?revision#mux.go
        const locationsByURI = new Map<string, Location[]>()

        // URIs with >0 locations, in order (to avoid jitter as more results stream in).
        const orderedURIs: { uri: string; repo: string }[] = []

        if (this.state.locationsOrError) {
            for (const location of this.state.locationsOrError) {
                if (!locationsByURI.has(location.uri)) {
                    locationsByURI.set(location.uri, [])

                    const { repoName } = parseRepoURI(location.uri)
                    orderedURIs.push({ uri: location.uri, repo: repoName })
                }
                locationsByURI.get(location.uri)!.push(location)
            }
        }

        return (
            <div className={classNames(styles.fileLocations, this.props.className)}>
                <VirtualList<OrderedURI, { locationsByURI: Map<string, Location[]> }>
                    itemsToShow={this.state.itemsToShow}
                    onShowMoreItems={this.onShowMoreItems}
                    items={orderedURIs}
                    renderItem={(
                        item: OrderedURI,
                        index: number,
                        additionalProps: { locationsByURI: Map<string, Location[]> }
                    ) => this.renderFileMatch(item, additionalProps, index)}
                    itemProps={{ locationsByURI }}
                    itemKey={this.itemKey}
                />
            </div>
        )
    }

    private onShowMoreItems = (): void => {
        this.setState(state => ({ itemsToShow: state.itemsToShow + 3 }))
    }

    private onSelect = (): void => {
        if (this.props.onSelect) {
            this.props.onSelect()
        }
    }

    private itemKey = (item: OrderedURI): string => item.uri

    private renderFileMatch = (
        { uri }: OrderedURI,
        { locationsByURI }: { locationsByURI: Map<string, Location[]> },
        index: number
    ): JSX.Element => (
        <FileContentSearchResult
            index={index}
            location={this.props.location}
            telemetryService={this.props.telemetryService}
            telemetryRecorder={this.props.telemetryRecorder}
            defaultExpanded={true}
            result={referencesToContentMatch(uri, locationsByURI.get(uri)!)}
            onSelect={this.onSelect}
            showAllMatches={true}
            fetchHighlightedFileLineRanges={this.props.fetchHighlightedFileLineRanges}
            settingsCascade={this.props.settingsCascade}
            containerClassName={styles.resultContainer}
        />
    )
}

function referencesToContentMatch(uri: string, references: Badged<Location>[]): ContentMatch {
    const parsedUri = parseRepoURI(uri)
    return {
        type: 'content',
        path: parsedUri.filePath || '',
        commit: (parsedUri.commitID || parsedUri.revision)!,
        repository: parsedUri.repoName,
        chunkMatches: references.filter(property('range', isDefined)).map(reference => ({
            content: '',
            contentStart: {
                line: reference.range.start.line,
                offset: 0,
                column: 0,
            },
            ranges: [
                {
                    start: {
                        line: reference.range.start.line,
                        offset: 0,
                        column: reference.range.start.character,
                    },
                    end: {
                        line: reference.range.end.line,
                        offset: 0,
                        column: reference.range.end.character,
                    },
                },
            ],
            aggregableBadges: reference.aggregableBadges,
        })),
    }
}
