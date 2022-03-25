import React, { useCallback, useEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { capitalize, find } from 'lodash'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import OpenInAppIcon from 'mdi-react/OpenInAppIcon'
import { MemoryRouter, useHistory, useLocation } from 'react-router'
import { Collapse } from 'reactstrap'

import { HoveredToken } from '@sourcegraph/codeintellify'
import {
    addLineRangeQueryParameter,
    formatSearchParameters,
    lprToRange,
    toPositionOrRangeQueryParameter,
    toViewStateHash,
} from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { getModeFromPath } from '@sourcegraph/shared/src/languages'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
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
    Link,
    LoadingSpinner,
    CardHeader,
    useDebounce,
    Button,
    Input,
    Icon,
    Badge,
    useObservable,
} from '@sourcegraph/wildcard'

import { ReferencesPanelHighlightedBlobResult, ReferencesPanelHighlightedBlobVariables } from '../graphql-operations'
import { resolveRevision } from '../repo/backend'
import { fetchBlob } from '../repo/blob/backend'
import { Blob } from '../repo/blob/Blob'
import { HoverThresholdProps } from '../repo/RepoContainer'
import { parseBrowserRepoURL } from '../util/url'

import { findLanguageSpec } from './language-specs/languages'
import { LanguageSpec } from './language-specs/spec'
import { Location, RepoLocationGroup, LocationGroup } from './location'
import { FETCH_HIGHLIGHTED_BLOB } from './ReferencesPanelQueries'
import { findSearchToken } from './token'
import { useCodeIntel } from './useCodeIntel'
import { isDefined } from './util/helpers'

import styles from './ReferencesPanel.module.scss'

type Token = HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec

interface ReferencesPanelProps
    extends SettingsCascadeProps,
        PlatformContextProps,
        TelemetryProps,
        HoverThresholdProps,
        ExtensionsControllerProps,
        ThemeProps {
    /** Whether to show the first loaded reference in mini code view */
    jumpToFirst?: boolean

    /**
     * The panel runs inside its own MemoryRouter, we keep track of externalHistory
     * so that we're still able to actually navigate within the browser when required
     */
    externalHistory: H.History
    externalLocation: H.Location
}

export const ReferencesPanelWithMemoryRouter: React.FunctionComponent<ReferencesPanelProps> = props => (
    <MemoryRouter
        // Force router to remount the Panel when external location changes
        key={`${props.externalLocation.pathname}${props.externalLocation.search}${props.externalLocation.hash}`}
        initialEntries={[props.externalLocation]}
    >
        <ReferencesPanel {...props} />
    </MemoryRouter>
)

const ReferencesPanel: React.FunctionComponent<ReferencesPanelProps> = props => {
    const location = useLocation()

    const { hash, pathname, search } = location
    const { line, character } = parseQueryAndHash(search, hash)
    const { filePath, repoName, ...parsedURL } = parseBrowserRepoURL(pathname)

    const revision = parsedURL.revision
    const commitID = parsedURL.commitID

    // If we don't have enough information in the URL, we can't render the panel
    if (!(line && character && filePath)) {
        return null
    }

    const searchParameters = new URLSearchParams(search)
    const jumpToFirst = searchParameters.get('jumpToFirst') === 'true'

    const token = { repoName, line, character, filePath }

    if (commitID === undefined || revision === undefined) {
        return <RevisionResolvingReferencesList {...props} {...token} jumpToFirst={jumpToFirst} />
    }

    return <FilterableReferencesList {...props} token={{ ...token, revision, commitID }} jumpToFirst={jumpToFirst} />
}

export const RevisionResolvingReferencesList: React.FunctionComponent<
    ReferencesPanelProps & {
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

    return <FilterableReferencesList {...props} token={token} />
}

interface ReferencesPanelPropsWithToken extends ReferencesPanelProps {
    token: Token
}

const FilterableReferencesList: React.FunctionComponent<ReferencesPanelPropsWithToken> = props => {
    const [filter, setFilter] = useState<string>()
    const debouncedFilter = useDebounce(filter, 150)

    useEffect(() => {
        setFilter(undefined)
    }, [props.token])

    const blobInfo = useObservable(
        fetchBlob({
            repoName: props.token.repoName,
            commitID: props.token.commitID,
            filePath: props.token.filePath,
            disableTimeout: false,
        })
    )

    const languageId = getModeFromPath(props.token.filePath)
    const spec = findLanguageSpec(languageId)
    const tokenResult = findSearchToken({
        text: blobInfo?.content ?? '',
        position: {
            line: props.token.line - 1,
            character: props.token.character - 1,
        },
        lineRegexes: spec.commentStyles.map(style => style.lineRegex).filter(isDefined),
        blockCommentStyles: spec.commentStyles.map(style => style.block).filter(isDefined),
        identCharPattern: spec.identCharPattern,
    })

    if (!tokenResult?.searchToken) {
        return (
            <div>
                <p className="text-danger">Could not find hovered token.</p>
            </div>
        )
    }

    return (
        <>
            <CardHeader>
                <code>{tokenResult.searchToken}</code>
            </CardHeader>
            <Input
                className={classNames('py-0 my-0', styles.referencesFilter)}
                type="text"
                placeholder="Filter by filename..."
                value={filter === undefined ? '' : filter}
                onChange={event => setFilter(event.target.value)}
            />
            <ReferencesList
                {...props}
                token={props.token}
                filter={debouncedFilter}
                searchToken={tokenResult?.searchToken}
                spec={spec}
            />
        </>
    )
}

const SHOW_SPINNER_DELAY_MS = 100

export const ReferencesList: React.FunctionComponent<
    ReferencesPanelPropsWithToken & {
        filter?: string
        searchToken: string
        spec: LanguageSpec
    }
> = props => {
    const {
        data,
        error,
        loading,
        referencesHasNextPage,
        implementationsHasNextPage,
        fetchMoreReferences,
        fetchMoreImplementations,
        fetchMoreReferencesLoading,
        fetchMoreImplementationsLoading,
    } = useCodeIntel({
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
        searchToken: props.searchToken,
        spec: props.spec,
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

    const references = useMemo(() => data?.references.nodes ?? [], [data])
    const definitions = useMemo(() => data?.definitions.nodes ?? [], [data])
    const implementations = useMemo(() => data?.implementations.nodes ?? [], [data])

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

    if (loading && !data) {
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
    if (error && !data) {
        return (
            <div>
                <p className="text-danger">Loading precise code intel failed:</p>
                <pre>{error.message}</pre>
            </div>
        )
    }

    // If there weren't any errors and we just didn't receive any data
    if (!data) {
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
                        <CardHeader className={classNames('pl-1 pr-3 py-1 d-flex justify-content-between')}>
                            <h4 className="mb-0">
                                {activeLocation.resource.path}{' '}
                                <Link
                                    to={activeLocation.url}
                                    onClick={event => {
                                        event.preventDefault()
                                        props.externalHistory.push(activeLocation.url)
                                    }}
                                >
                                    <Icon as={OpenInAppIcon} />
                                </Link>
                            </h4>

                            <Button
                                onClick={() => setActiveLocation(undefined)}
                                className={classNames('btn-icon p-0', styles.dismissButton)}
                                title="Close panel"
                                data-tooltip="Close panel"
                                data-placement="left"
                            >
                                <Icon as={CloseIcon} />
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

interface CollapsibleLocationListProps {
    name: string
    locations: Location[]
    setActiveLocation: (location: Location | undefined) => void
    activeLocation: Location | undefined
    filter: string | undefined
    hasMore: boolean
    fetchMore?: () => void
    loadingMore: boolean
}

const CollapsibleLocationList: React.FunctionComponent<CollapsibleLocationListProps> = props => {
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
                            <Icon aria-label="Close" as={ChevronDownIcon} />
                        ) : (
                            <Icon aria-label="Expand" as={ChevronRightIcon} />
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
            </Collapse>
        </>
    )
}

const SideBlob: React.FunctionComponent<
    ReferencesPanelProps & {
        activeLocation: Location

        location: H.Location
        history: H.History
        blobNav: (url: string) => void
    }
> = props => {
    const { data, error, loading } = useQuery<
        ReferencesPanelHighlightedBlobResult,
        ReferencesPanelHighlightedBlobVariables
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

interface LocationsListProps {
    locations: Location[]
    activeLocation?: Location
    setActiveLocation: (reference: Location | undefined) => void
    filter: string | undefined
}

const LocationsList: React.FunctionComponent<LocationsListProps> = ({
    locations,
    activeLocation,
    setActiveLocation,
    filter,
}) => {
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
                <CollapsibleRepoLocationGroup
                    key={repoReferenceGroup.repoName}
                    repoLocationGroup={repoReferenceGroup}
                    activeLocation={activeLocation}
                    setActiveLocation={setActiveLocation}
                    getLineContent={getLineContent}
                    filter={filter}
                />
            ))}
        </>
    )
}

const CollapsibleRepoLocationGroup: React.FunctionComponent<{
    repoLocationGroup: RepoLocationGroup
    activeLocation?: Location
    setActiveLocation: (reference: Location | undefined) => void
    getLineContent: (location: Location) => string
    filter: string | undefined
}> = ({ repoLocationGroup, setActiveLocation, getLineContent, activeLocation, filter }) => {
    const [isOpen, setOpen] = useState<boolean>(true)
    const handleOpen = useCallback(() => setOpen(previousState => !previousState), [])

    const allSearchBased = useMemo(
        () =>
            find(
                repoLocationGroup.referenceGroups.flatMap(reference => reference.locations),
                location => !location.precise
            ) !== undefined,
        [repoLocationGroup]
    )

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
                        <Icon aria-label="Close" as={ChevronDownIcon} />
                    ) : (
                        <Icon aria-label="Expand" as={ChevronRightIcon} />
                    )}

                    <Link to={`/${repoLocationGroup.repoName}`}>{displayRepoName(repoLocationGroup.repoName)}</Link>

                    <Badge pill={true} small={true} variant="secondary" className="ml-2">
                        {allSearchBased ? 'SEARCH-BASED' : 'PRECISE'}
                    </Badge>
                </span>
            </Button>

            <Collapse id={repoLocationGroup.repoName} isOpen={isOpen}>
                {repoLocationGroup.referenceGroups.map(group => (
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
                <span className={styles.referenceFilename}>
                    {isOpen ? (
                        <Icon aria-label="Close" as={ChevronDownIcon} />
                    ) : (
                        <Icon aria-label="Expand" as={ChevronRightIcon} />
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
                            activeLocation && activeLocation.url === reference.url ? styles.referenceActive : ''

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
