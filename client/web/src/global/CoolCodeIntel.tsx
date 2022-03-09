import classNames from 'classnames'
import * as H from 'history'
import { capitalize } from 'lodash'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import OpenInAppIcon from 'mdi-react/OpenInAppIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { MemoryRouter, useHistory, useLocation } from 'react-router'
import { Collapse } from 'reactstrap'

import { HoveredToken } from '@sourcegraph/codeintellify'
import {
    addLineRangeQueryParameter,
    formatSearchParameters,
    isErrorLike,
    lprToRange,
    toPositionOrRangeQueryParameter,
    toViewStateHash,
} from '@sourcegraph/common'
import { Range } from '@sourcegraph/extension-api-types'
import { useQuery } from '@sourcegraph/http-client'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { Resizable } from '@sourcegraph/shared/src/components/Resizable'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeOrError, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import {
    RepoSpec,
    RevisionSpec,
    FileSpec,
    ResolvedRevisionSpec,
    parseQueryAndHash,
} from '@sourcegraph/shared/src/util/url'
import {
    Tab,
    TabList,
    TabPanel,
    TabPanels,
    Tabs,
    Link,
    LoadingSpinner,
    useLocalStorage,
    CardHeader,
    useDebounce,
    Button,
    useObservable,
    Input,
    Badge,
} from '@sourcegraph/wildcard'

import { ErrorBoundary } from '../components/ErrorBoundary'
import {
    CoolCodeIntelHighlightedBlobResult,
    CoolCodeIntelHighlightedBlobVariables,
    LocationFields,
} from '../graphql-operations'
import { resolveRevision } from '../repo/backend'
import { Blob } from '../repo/blob/Blob'
import { HoverThresholdProps } from '../repo/RepoContainer'
import { parseBrowserRepoURL } from '../util/url'

import styles from './CoolCodeIntel.module.scss'
import { FETCH_HIGHLIGHTED_BLOB } from './CoolCodeIntelQueries'
import { usePreciseCodeIntel } from './usePreciseCodeIntel'

export interface GlobalCoolCodeIntelProps {
    coolCodeIntelEnabled: boolean
}

type Token = HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec

interface CoolCodeIntelProps
    extends SettingsCascadeProps,
        PlatformContextProps,
        TelemetryProps,
        HoverThresholdProps,
        ExtensionsControllerProps,
        ThemeProps {
    // The token for which to show references
    token?: Token

    jumpToFirst?: boolean
    /**
     * The panel runs inside its own MemoryRouter, we keep track of externalHistory
     * so that we're still able to actually navigate within the browser when required
     */
    externalHistory: H.History
    externalLocation: H.Location
}

export const isCoolCodeIntelEnabled = (settingsCascade: SettingsCascadeOrError): boolean =>
    !isErrorLike(settingsCascade.final) && settingsCascade.final?.experimentalFeatures?.coolCodeIntel === true

export const CoolCodeIntel: React.FunctionComponent<CoolCodeIntelProps> = props => (
    <ErrorBoundary
        location={null}
        render={error => (
            <div>
                <pre>{JSON.stringify(error)}</pre>
            </div>
        )}
    >
        <MemoryRouter
            // Force router to remount the Panel when external location changes
            key={`${props.externalLocation.pathname}${props.externalLocation.search}${props.externalLocation.hash}`}
            initialEntries={[props.externalLocation]}
        >
            <CoolCodeIntelResizablePanel {...props} />
        </MemoryRouter>
    </ErrorBoundary>
)

const LAST_TAB_STORAGE_KEY = 'CoolCodeIntel.lastTab'

type CoolCodeIntelTabID = 'references' | 'token' | 'definition'

interface CoolCodeIntelTab {
    id: CoolCodeIntelTabID
    label: string
    component: React.ComponentType<CoolCodeIntelProps>
}

export const ReferencesPanel: React.FunctionComponent<CoolCodeIntelProps> = props => {
    if (!props.token) {
        return null
    }

    return <FilterableReferencesList token={props.token} {...props} />
}

interface Location {
    resource: {
        path: string
        content: string
        repository: {
            name: string
        }
        commit: {
            oid: string
        }
    }
    range?: Range
    url: string
    lines: string[]
}

const buildLocation = (node: LocationFields): Location => {
    const location: Location = {
        resource: {
            repository: { name: node.resource.repository.name },
            content: node.resource.content,
            path: node.resource.path,
            commit: node.resource.commit,
        },
        url: '',
        lines: [],
    }
    if (node.range !== null) {
        location.range = node.range
    }
    location.url = node.url
    location.lines = location.resource.content.split(/\r?\n/)
    return location
}

interface RepoLocationGroup {
    repoName: string
    referenceGroups: LocationGroup[]
}

interface LocationGroup {
    repoName: string
    path: string
    locations: Location[]
}

const FilterableReferencesList: React.FunctionComponent<
    CoolCodeIntelProps & {
        token: Token
    }
> = props => {
    const [filter, setFilter] = useState<string>()
    const debouncedFilter = useDebounce(filter, 150)

    useEffect(() => {
        setFilter(undefined)
    }, [props.token])

    return (
        <>
            <Input
                className={classNames('py-0 my-0', styles.referencesFilter)}
                type="text"
                placeholder="Filter by filename..."
                value={filter === undefined ? '' : filter}
                onChange={event => setFilter(event.target.value)}
            />
            <ReferencesList {...props} filter={debouncedFilter} />
        </>
    )
}

const SHOW_SPINNER_DELAY_MS = 100

export const ReferencesList: React.FunctionComponent<
    CoolCodeIntelProps & {
        token: Token
        filter?: string
    }
> = props => {
    const {
        lsifData,
        error,
        loading,
        referencesHasNextPage,
        implementationsHasNextPage,
        fetchMoreReferences,
        fetchMoreImplementations,
        fetchMoreReferencesLoading,
        fetchMoreImplementationsLoading,
    } = usePreciseCodeIntel({
        variables: {
            repository: props.token.repoName,
            commit: props.token.commitID,
            path: props.token.filePath,
            // On the backend the line/character are 0-indexed, but what we
            // get from hoverifier is 1-indexed.
            line: props.token.line - 1,
            character: props.token.character - 1,
            filter: props.filter || null,
            firstReferences: 100,
            afterReferences: null,
            firstImplementations: 100,
            afterImplementations: null,
        },
    })

    // We only show the inline loading message if loading takes longer than
    // SHOW_SPINNER_DELAY_MS milliseconds.
    const [canShowSpinner, setCanShowSpinner] = useState(false)
    useEffect(() => {
        const handle = setTimeout(() => setCanShowSpinner(loading), SHOW_SPINNER_DELAY_MS)
        // In case the component un-mounts before
        return () => clearTimeout(handle)
        // Make sure this effect only runs once
    }, [loading])

    const references = useMemo(() => (lsifData?.references.nodes ?? []).map(buildLocation), [lsifData])
    const definitions = useMemo(() => (lsifData?.definitions.nodes ?? []).map(buildLocation), [lsifData])
    const implementations = useMemo(() => (lsifData?.implementations.nodes ?? []).map(buildLocation), [lsifData])

    // activeLocation is the location that is selected/clicked in the list of
    // definitions/references/implementations.
    const [activeLocation, setActiveLocation] = useState<Location>()
    // We create an in-memory history here so we don't modify the browser
    // location. This panel is detached from the URL state.
    const blobMemoryHistory = useMemo(() => H.createMemoryHistory(), [])

    // When the token for which we display data changed, we want to reset
    // activeLocation.
    // But only if we are not re-rendering with different token and the code
    // blob already open.
    useEffect(() => {
        if (!props.jumpToFirst) {
            setActiveLocation(undefined)
        }
    }, [props.jumpToFirst, props.token])

    // If props.jumpToFirst is true and we finished loading (and have
    // definitions) we select the first definition. We set it as activeLocation
    // and push it to the blobMemoryHistory so the code blob is open.
    useEffect(() => {
        if (props.jumpToFirst && definitions.length > 0) {
            blobMemoryHistory.push(definitions[0].url)
            setActiveLocation(definitions[0])
        }
    }, [blobMemoryHistory, props.jumpToFirst, definitions])

    // When a user clicks on an item in the list of references, we push it to
    // the memory history for the code blob on the right, so it will jump to &
    // highlight the correct line.
    const onReferenceClick = (location: Location | undefined): void => {
        if (location) {
            blobMemoryHistory.push(location.url)
        }
        setActiveLocation(location)
    }

    // This is the history of the panel, that is inside a memory router
    const panelHistory = useHistory()
    // When we user clicks on a token *inside* the code blob on the right, we
    // update the history for the panel itself, which is inside a memory router.
    //
    // We also '#tab=references' and '?jumpToFirst=true' to the URL.
    //
    // '#tab=references' will cause the panel to show the references of the clicked token,
    // but not navigate the main web app to it.
    //
    // '?jumpToFirst=true' causes the panel to select the first reference and
    // open it in code blob on right.
    const onBlobNav = (url: string): void => {
        // If we're going to navigate inside the same file in the same repo we
        // can optimistically jump to that position in the code blob.
        const resource = activeLocation?.resource
        if (resource !== undefined) {
            const urlToken = tokenFromUrl(url)
            if (urlToken.filePath === resource.path && urlToken.repoName === resource.repository.name) {
                blobMemoryHistory.push(url)
            }
        }

        panelHistory.push(appendJumpToFirstQueryParameter(url) + toViewStateHash('references'))
    }

    if (loading && !lsifData) {
        return (
            <>
                <LoadingSpinner inline={false} className="mx-auto my-4" />
                <p className="text-muted text-center">
                    <i>Loading precise code intel ...</i>
                </p>
            </>
        )
    }

    // If we received an error before we had received any data
    if (error && !lsifData) {
        return (
            <div>
                <p className="text-danger">Loading precise code intel failed:</p>
                <pre>{error.message}</pre>
            </div>
        )
    }

    // If there weren't any errors and we just didn't receive any data
    if (!lsifData) {
        return <>Nothing found</>
    }

    return (
        <>
            <div className={classNames('align-items-stretch', styles.referencesList)}>
                <div className={classNames('px-0', styles.referencesSideReferences)}>
                    {canShowSpinner && (
                        <div className="text-muted">
                            <LoadingSpinner inline={true} />
                            <i>Loading...</i>
                        </div>
                    )}
                    <CollapsibleLocationList
                        {...props}
                        name="definitions"
                        locations={definitions}
                        hasMore={false}
                        loadingMore={false}
                        setActiveLocation={onReferenceClick}
                        filter={props.filter}
                        activeLocation={activeLocation}
                    />
                    <CollapsibleLocationList
                        {...props}
                        name="references"
                        locations={references}
                        hasMore={referencesHasNextPage}
                        fetchMore={fetchMoreReferences}
                        loadingMore={fetchMoreReferencesLoading}
                        setActiveLocation={onReferenceClick}
                        filter={props.filter}
                        activeLocation={activeLocation}
                    />
                    {implementations.length > 0 && (
                        <CollapsibleLocationList
                            {...props}
                            name="implementations"
                            locations={implementations}
                            hasMore={implementationsHasNextPage}
                            fetchMore={fetchMoreImplementations}
                            loadingMore={fetchMoreImplementationsLoading}
                            setActiveLocation={onReferenceClick}
                            filter={props.filter}
                            activeLocation={activeLocation}
                        />
                    )}
                </div>
                {activeLocation !== undefined && (
                    <div className={classNames('px-0 border-left', styles.referencesSideBlob)}>
                        <CardHeader
                            className={classNames(
                                'pl-1 d-flex justify-content-between',
                                styles.referencesSideBlobFilename
                            )}
                        >
                            <h4>
                                {activeLocation.resource.path}{' '}
                                <Link
                                    to={activeLocation.url}
                                    onClick={event => {
                                        event.preventDefault()
                                        props.externalHistory.push(activeLocation.url)
                                    }}
                                >
                                    <OpenInAppIcon className="icon-inline" />
                                </Link>
                            </h4>

                            <Button
                                onClick={() => setActiveLocation(undefined)}
                                className={classNames('btn-icon py-0', styles.dismissButton)}
                                title="Close panel"
                                data-tooltip="Close panel"
                                data-placement="left"
                            >
                                <CloseIcon className="icon-inline" />
                            </Button>
                        </CardHeader>
                        <SideBlob
                            {...props}
                            blobNav={onBlobNav}
                            history={blobMemoryHistory}
                            location={blobMemoryHistory.location}
                            activeLocation={activeLocation}
                        />
                    </div>
                )}
            </div>
        </>
    )
}

const CollapsibleLocationList: React.FunctionComponent<{
    name: string
    locations: Location[]
    setActiveLocation: (location: Location | undefined) => void
    activeLocation: Location | undefined
    filter: string | undefined
    hasMore: boolean
    fetchMore?: () => void
    loadingMore: boolean
}> = props => {
    const [isOpen, setOpen] = useState<boolean>(true)
    const handleOpen = useCallback(() => setOpen(previousState => !previousState), [])

    return (
        <>
            <CardHeader className="p-0">
                <Button
                    aria-expanded={isOpen}
                    type="button"
                    onClick={handleOpen}
                    className="bg-transparent py-1 px-0 border-bottom border-top-0 border-left-0 border-right-0 d-flex justify-content-start w-100"
                >
                    <h4 className="px-1 py-0 mb-0">
                        {' '}
                        {isOpen ? (
                            <ChevronDownIcon className="icon-inline" aria-label="Close" />
                        ) : (
                            <ChevronRightIcon className="icon-inline" aria-label="Expand" />
                        )}{' '}
                        {capitalize(props.name)}
                        <Badge pill={true} variant="secondary" className="ml-2">
                            {props.locations.length}
                            {props.hasMore && '+'}
                        </Badge>
                    </h4>
                </Button>
            </CardHeader>

            <Collapse id="references" isOpen={isOpen}>
                {props.locations.length > 0 ? (
                    <>
                        <LocationsList
                            locations={props.locations}
                            activeLocation={props.activeLocation}
                            setActiveLocation={props.setActiveLocation}
                            filter={props.filter}
                        />
                        {props.hasMore &&
                            props.fetchMore !== undefined &&
                            (props.loadingMore ? (
                                <div className="text-center mb-1">
                                    <em>Loading more {props.name}...</em>
                                    <LoadingSpinner inline={true} />
                                </div>
                            ) : (
                                <div className="text-center mb-1">
                                    <Button variant="secondary" onClick={props.fetchMore}>
                                        Load more {props.name}
                                    </Button>
                                </div>
                            ))}
                    </>
                ) : (
                    <p className="text-muted pl-2">
                        {props.filter ? (
                            <i>
                                No {props.name} matching <strong>{props.filter}</strong> found
                            </i>
                        ) : (
                            <i>No {props.name} found</i>
                        )}
                    </p>
                )}
            </Collapse>
        </>
    )
}

const SideBlob: React.FunctionComponent<
    CoolCodeIntelProps & {
        activeLocation: Location

        location: H.Location
        history: H.History
        blobNav: (url: string) => void
    }
> = props => {
    const { data, error, loading } = useQuery<
        CoolCodeIntelHighlightedBlobResult,
        CoolCodeIntelHighlightedBlobVariables
    >(FETCH_HIGHLIGHTED_BLOB, {
        variables: {
            repository: props.activeLocation.resource.repository.name,
            commit: props.activeLocation.resource.commit.oid,
            path: props.activeLocation.resource.path,
        },
        // Cache this data but always re-request it in the background when we revisit
        // this page to pick up newer changes.
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'network-only',
    })

    // If we're loading and haven't received any data yet
    if (loading && !data) {
        return (
            <>
                <LoadingSpinner inline={false} className="mx-auto my-4" />
                <p className="text-muted text-center">
                    <i>
                        Loading <code>{props.activeLocation.resource.path}</code>...
                    </i>
                </p>
            </>
        )
    }

    // If we received an error before we had received any data
    if (error && !data) {
        return (
            <div>
                <p className="text-danger">
                    Loading <code>{props.activeLocation.resource.path}</code> failed:
                </p>
                <pre>{error.message}</pre>
            </div>
        )
    }

    // If there weren't any errors and we just didn't receive any data
    if (!data?.repository?.commit?.blob?.highlight) {
        return <>Nothing found</>
    }

    const { html, aborted } = data?.repository?.commit?.blob?.highlight
    if (aborted) {
        return (
            <p className="text-warning text-center">
                <i>
                    Highlighting <code>{props.activeLocation.resource.path}</code> failed
                </i>
            </p>
        )
    }

    return (
        <Blob
            {...props}
            nav={props.blobNav}
            history={props.history}
            location={props.location}
            disableStatusBar={true}
            wrapCode={true}
            className={styles.referencesSideBlobCode}
            blobInfo={{
                content: props.activeLocation.resource.content,
                html,
                filePath: props.activeLocation.resource.path,
                repoName: props.activeLocation.resource.repository.name,
                commitID: props.activeLocation.resource.commit.oid,
                revision: props.activeLocation.resource.commit.oid,
                mode: 'lspmode',
            }}
        />
    )
}

const getLineContent = (location: Location): string => {
    const range = location.range
    if (range !== undefined) {
        return location.lines[range.start?.line].trim()
    }
    return ''
}

const LocationsList: React.FunctionComponent<{
    locations: Location[]
    activeLocation?: Location
    setActiveLocation: (reference: Location | undefined) => void
    filter: string | undefined
}> = ({ locations, activeLocation, setActiveLocation, filter }) => {
    const repoLocationGroups = useMemo((): RepoLocationGroup[] => {
        const byFile: Record<string, Location[]> = {}
        for (const location of locations) {
            if (byFile[location.resource.path] === undefined) {
                byFile[location.resource.path] = []
            }
            byFile[location.resource.path].push(location)
        }

        const locationsGroups: LocationGroup[] = []
        Object.keys(byFile).map(path => {
            const references = byFile[path]
            const repoName = references[0].resource.repository.name
            locationsGroups.push({ path, locations: references, repoName })
        })

        const byRepo: Record<string, LocationGroup[]> = {}
        for (const group of locationsGroups) {
            if (byRepo[group.repoName] === undefined) {
                byRepo[group.repoName] = []
            }
            byRepo[group.repoName].push(group)
        }
        const repoLocationGroups: RepoLocationGroup[] = []
        Object.keys(byRepo).map(repoName => {
            const referenceGroups = byRepo[repoName]
            repoLocationGroups.push({ repoName, referenceGroups })
        })
        return repoLocationGroups
    }, [locations])

    return (
        <>
            {repoLocationGroups.map(repoReferenceGroup => (
                <RepoReferenceGroup
                    key={repoReferenceGroup.repoName}
                    repoReferenceGroup={repoReferenceGroup}
                    activeLocation={activeLocation}
                    setActiveLocation={setActiveLocation}
                    getLineContent={getLineContent}
                    filter={filter}
                />
            ))}
        </>
    )
}

const RepoReferenceGroup: React.FunctionComponent<{
    repoReferenceGroup: RepoLocationGroup
    activeLocation?: Location
    setActiveLocation: (reference: Location | undefined) => void
    getLineContent: (location: Location) => string
    filter: string | undefined
}> = ({ repoReferenceGroup, setActiveLocation, getLineContent, activeLocation, filter }) => {
    const [isOpen, setOpen] = useState<boolean>(true)
    const handleOpen = useCallback(() => setOpen(previousState => !previousState), [])

    return (
        <>
            <Button
                aria-expanded={isOpen}
                type="button"
                onClick={handleOpen}
                className="bg-transparent py-1 border-bottom border-top-0 border-left-0 border-right-0 d-flex justify-content-start w-100"
            >
                <span className="p-0 mb-0">
                    {isOpen ? (
                        <ChevronDownIcon className="icon-inline" aria-label="Close" />
                    ) : (
                        <ChevronRightIcon className="icon-inline" aria-label="Expand" />
                    )}

                    <Link to={`/${repoReferenceGroup.repoName}`}>{displayRepoName(repoReferenceGroup.repoName)}</Link>
                </span>
            </Button>

            <Collapse id={repoReferenceGroup.repoName} isOpen={isOpen}>
                {repoReferenceGroup.referenceGroups.map(group => (
                    <ReferenceGroup
                        key={group.path + group.repoName}
                        group={group}
                        activeLocation={activeLocation}
                        setActiveLocation={setActiveLocation}
                        getLineContent={getLineContent}
                        filter={filter}
                    />
                ))}
            </Collapse>
        </>
    )
}

const ReferenceGroup: React.FunctionComponent<{
    group: LocationGroup
    activeLocation?: Location
    setActiveLocation: (reference: Location | undefined) => void
    getLineContent: (reference: Location) => string
    filter: string | undefined
}> = ({ group, setActiveLocation: setActiveLocation, getLineContent, activeLocation, filter }) => {
    const [isOpen, setOpen] = useState<boolean>(true)
    const handleOpen = useCallback(() => setOpen(previousState => !previousState), [])

    let highlighted = [group.path]
    if (filter !== undefined) {
        highlighted = group.path.split(filter)
    }

    return (
        <div className="ml-4">
            <Button
                aria-expanded={isOpen}
                type="button"
                onClick={handleOpen}
                className="bg-transparent py-1 border-bottom border-top-0 border-left-0 border-right-0 d-flex justify-content-start w-100"
            >
                <span className={styles.coolCodeIntelReferenceFilename}>
                    {isOpen ? (
                        <ChevronDownIcon className="icon-inline" aria-label="Close" />
                    ) : (
                        <ChevronRightIcon className="icon-inline" aria-label="Expand" />
                    )}
                    {highlighted.length === 2 ? (
                        <span>
                            {highlighted[0]}
                            <mark>{filter}</mark>
                            {highlighted[1]}
                        </span>
                    ) : (
                        group.path
                    )}{' '}
                    ({group.locations.length} references)
                </span>
            </Button>

            <Collapse id={group.repoName + group.path} isOpen={isOpen} className="ml-2">
                <ul className="list-unstyled pl-3 py-1 mb-0">
                    {group.locations.map(reference => {
                        const className =
                            activeLocation && activeLocation.url === reference.url
                                ? styles.coolCodeIntelReferenceActive
                                : ''

                        return (
                            <li key={reference.url} className={classNames('border-0 rounded-0', className)}>
                                <div>
                                    <Link
                                        onClick={event => {
                                            event.preventDefault()
                                            setActiveLocation(reference)
                                        }}
                                        to={reference.url}
                                        className={styles.referenceLink}
                                    >
                                        <span className={styles.referenceLinkLineNumber}>
                                            {(reference.range?.start?.line ?? 0) + 1}
                                            {': '}
                                        </span>
                                        <code>{getLineContent(reference)}</code>
                                    </Link>
                                </div>
                            </li>
                        )
                    })}
                </ul>
            </Collapse>
        </div>
    )
}

const TABS: CoolCodeIntelTab[] = [{ id: 'references', label: 'References', component: ReferencesPanel }]

const ResizableCoolCodeIntelPanel = React.memo<CoolCodeIntelProps>(props => (
    <Resizable
        className={styles.resizablePanel}
        handlePosition="top"
        defaultSize={350}
        storageKey="panel-size"
        element={<CoolCodeIntelPanel {...props} />}
    />
))

const CoolCodeIntelPanel = React.memo<CoolCodeIntelProps>(props => {
    const [tabIndex, setTabIndex] = useLocalStorage(LAST_TAB_STORAGE_KEY, 0)
    const handleTabsChange = useCallback((index: number) => setTabIndex(index), [setTabIndex])

    const location = useLocation()
    const handlePanelClose = useCallback(() => {
        // We close the panel by removing the viewState in the external history
        props.externalHistory.push(locationWithoutViewState(location))
    }, [props.externalHistory, location])

    return (
        <Tabs size="medium" className={styles.panel} index={tabIndex} onChange={handleTabsChange}>
            <div
                className={classNames('tablist-wrapper d-flex justify-content-between sticky-top', styles.panelHeader)}
            >
                <TabList>
                    <div className="d-flex w-100">
                        {TABS.map(({ label, id }) => (
                            <Tab key={id}>
                                <span className="tablist-wrapper--tab-label" role="none">
                                    {label}
                                </span>
                            </Tab>
                        ))}
                    </div>
                </TabList>
                <div className="align-items-center d-flex">
                    <Button
                        onClick={handlePanelClose}
                        className={classNames('btn-icon ml-2', styles.dismissButton)}
                        title="Close panel"
                        data-tooltip="Close panel"
                        data-placement="left"
                    >
                        <CloseIcon className="icon-inline" />
                    </Button>
                </div>
            </div>
            <TabPanels>
                {TABS.map(tab => (
                    <TabPanel key={tab.id}>
                        <tab.component {...props} />
                    </TabPanel>
                ))}
            </TabPanels>
        </Tabs>
    )
})

export function locationWithoutViewState(location: H.Location): H.LocationDescriptorObject {
    const parsedQuery = parseQueryAndHash(location.search, location.hash)
    delete parsedQuery.viewState

    const lineRangeQueryParameter = toPositionOrRangeQueryParameter({ range: lprToRange(parsedQuery) })
    const result = {
        search: formatSearchParameters(
            addLineRangeQueryParameter(new URLSearchParams(location.search), lineRangeQueryParameter)
        ),
        hash: '',
    }
    return result
}

const CoolCodeIntelResizablePanel: React.FunctionComponent<CoolCodeIntelProps> = props => {
    const location = useLocation()

    const { hash, pathname, search } = location
    const { line, character, viewState } = parseQueryAndHash(search, hash)
    const { filePath, repoName, revision, commitID } = parseBrowserRepoURL(pathname)

    // If we don't have enough information in the URL, we can't render the panel
    if (!(line && character && filePath && viewState)) {
        return null
    }

    const searchParameters = new URLSearchParams(search)
    const jumpToFirst = searchParameters.get('jumpToFirst') === 'true'

    const token = { repoName, line, character, filePath }

    if (commitID === undefined || revision === undefined) {
        return <RevisionResolvingCoolCodeIntelPanel {...props} {...token} jumpToFirst={jumpToFirst} />
    }

    return <ResizableCoolCodeIntelPanel {...props} token={{ ...token, revision, commitID }} jumpToFirst={jumpToFirst} />
}

export const RevisionResolvingCoolCodeIntelPanel: React.FunctionComponent<
    CoolCodeIntelProps & {
        repoName: string
        line: number
        character: number
        filePath: string
        revision?: string
    }
> = props => {
    const resolvedRevision = useObservable(useMemo(() => resolveRevision(props), [props]))

    if (!resolvedRevision) {
        return null
    }

    const token = {
        repoName: props.repoName,
        line: props.line,
        character: props.character,
        filePath: props.filePath,

        revision: props.revision || resolvedRevision.defaultBranch,
        commitID: resolvedRevision.commitID,
    }

    return <ResizableCoolCodeIntelPanel {...props} token={token} />
}

export const appendJumpToFirstQueryParameter = (url: string): string => {
    const newUrl = new URL(url, window.location.href)
    newUrl.searchParams.set('jumpToFirst', 'true')
    return newUrl.pathname + `?${formatSearchParameters(newUrl.searchParams)}` + newUrl.hash
}

const tokenFromUrl = (url: string): { repoName: string; commitID?: string; filePath?: string } => {
    const parsed = new URL(url, window.location.href)

    const { filePath, repoName, commitID } = parseBrowserRepoURL(parsed.pathname)

    return { repoName, filePath, commitID }
}
