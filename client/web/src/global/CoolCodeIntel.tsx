import classNames from 'classnames'
import * as H from 'history'
import CloseIcon from 'mdi-react/CloseIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import OpenInAppIcon from 'mdi-react/OpenInAppIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useHistory, useLocation } from 'react-router'
import { Collapse } from 'reactstrap'

import { HoveredToken } from '@sourcegraph/codeintellify'
import {
    addLineRangeQueryParameter,
    appendLineRangeQueryParameter,
    appendSubtreeQueryParameter,
    formatSearchParameters,
    isErrorLike,
    lprToRange,
    renderMarkdown,
    toPositionOrRangeQueryParameter,
} from '@sourcegraph/common'
import { Range } from '@sourcegraph/extension-api-types'
import { useQuery } from '@sourcegraph/http-client'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { Resizable } from '@sourcegraph/shared/src/components/Resizable'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
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
} from '@sourcegraph/wildcard'

import { ErrorBoundary } from '../components/ErrorBoundary'
import {
    CoolCodeIntelHighlightedBlobResult,
    CoolCodeIntelHighlightedBlobVariables,
    CoolCodeIntelReferencesResult,
    CoolCodeIntelReferencesVariables,
    HoverFields,
    LocationConnectionFields,
    LocationFields,
    Maybe,
} from '../graphql-operations'
import { resolveRevision } from '../repo/backend'
import { Blob, BlobProps } from '../repo/blob/Blob'
import { parseBrowserRepoURL } from '../util/url'

import styles from './CoolCodeIntel.module.scss'
import { FETCH_HIGHLIGHTED_BLOB, FETCH_REFERENCES_QUERY } from './CoolCodeIntelQueries'

export interface GlobalCoolCodeIntelProps {
    coolCodeIntelEnabled: boolean
    onTokenClick?: (clickedToken: CoolClickedToken) => void
}

export type CoolClickedToken = HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec

interface CoolCodeIntelProps
    extends Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo' | 'disableStatusBar' | 'coolCodeIntelEnabled'> {
    clickedToken?: CoolClickedToken
}

export const isCoolCodeIntelEnabled = (settingsCascade: SettingsCascadeOrError): boolean =>
    !isErrorLike(settingsCascade.final) && settingsCascade.final?.experimentalFeatures?.coolCodeIntel === true

export const CoolCodeIntel: React.FunctionComponent<CoolCodeIntelProps & { onClose: () => void }> = props => (
    <ErrorBoundary
        location={null}
        render={error => (
            <div>
                <pre>{JSON.stringify(error)}</pre>
            </div>
        )}
    >
        <CoolCodeIntelResizablePanel {...props} />
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
    if (!props.clickedToken) {
        return null
    }

    return <ReferencesList clickedToken={props.clickedToken} {...props} />
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
    location.url = buildFileURL(location)
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

export const ReferencesList: React.FunctionComponent<
    CoolCodeIntelProps & {
        clickedToken: CoolClickedToken
    }
> = props => {
    const [activeLocation, setActiveLocation] = useState<Location>()
    const [filter, setFilter] = useState<string>()
    const debouncedFilter = useDebounce(filter, 150)

    useEffect(() => {
        setActiveLocation(undefined)
        setFilter(undefined)
    }, [props.clickedToken])

    // We create an in-memory history here so we don't modify the browser
    // location. This panel is detached from the URL state.
    const history = useMemo(() => H.createMemoryHistory(), [])

    const onReferenceClick = (location: Location | undefined): void => {
        if (location) {
            history.push(location.url)
        }
        setActiveLocation(location)
    }

    return (
        <>
            <Input
                className={classNames('py-0 my-0', styles.referencesFilter)}
                type="text"
                placeholder="Filter by filename..."
                value={filter === undefined ? '' : filter}
                onChange={event => setFilter(event.target.value)}
            />
            <div className={classNames('align-items-stretch', styles.referencesList)}>
                <div className={classNames('px-0', styles.referencesSideReferences)}>
                    <SideReferences
                        {...props}
                        activeLocation={activeLocation}
                        setActiveLocation={onReferenceClick}
                        filter={debouncedFilter}
                    />
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
                                <Link to={activeLocation.url}>
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
                            history={history}
                            location={history.location}
                            activeLocation={activeLocation}
                        />
                    </div>
                )}
            </div>
        </>
    )
}

export const SideReferences: React.FunctionComponent<
    CoolCodeIntelProps & {
        clickedToken: CoolClickedToken
        setActiveLocation: (location: Location | undefined) => void
        activeLocation: Location | undefined
        filter: string | undefined
    }
> = props => {
    const { data, error, loading } = useQuery<CoolCodeIntelReferencesResult, CoolCodeIntelReferencesVariables>(
        FETCH_REFERENCES_QUERY,
        {
            variables: {
                repository: props.clickedToken.repoName,
                commit: props.clickedToken.commitID,
                path: props.clickedToken.filePath,
                // On the backend the line/character are 0-indexed, but what we
                // get from hoverifier is 1-indexed.
                line: props.clickedToken.line - 1,
                character: props.clickedToken.character - 1,
                after: null,
                filter: props.filter || null,
            },
            // Cache this data but always re-request it in the background when we revisit
            // this page to pick up newer changes.
            fetchPolicy: 'cache-and-network',
            nextFetchPolicy: 'network-only',
        }
    )

    // If we're loading and haven't received any data yet
    if (loading && !data) {
        return (
            <>
                <LoadingSpinner inline={false} className="mx-auto my-4" />
                <p className="text-muted text-center">
                    <i>Loading references ...</i>
                </p>
            </>
        )
    }

    // If we received an error before we had received any data
    if (error && !data) {
        return (
            <div>
                <p className="text-danger">Loading references failed:</p>
                <pre>{error.message}</pre>
            </div>
        )
    }

    // If there weren't any errors and we just didn't receive any data
    if (!data || !data.repository?.commit?.blob?.lsif) {
        return <>Nothing found</>
    }

    const lsif = data.repository?.commit?.blob?.lsif

    return (
        <SideReferencesLists
            {...props}
            references={lsif.references}
            definitions={lsif.definitions}
            implementations={lsif.implementations}
            hover={lsif.hover}
        />
    )
}

const SideReferencesLists: React.FunctionComponent<
    CoolCodeIntelProps & {
        clickedToken: CoolClickedToken
        setActiveLocation: (location: Location | undefined) => void
        activeLocation: Location | undefined
        filter: string | undefined
        references: LocationConnectionFields
        definitions: Omit<LocationConnectionFields, 'pageInfo'>
        implementations: LocationConnectionFields
        hover: Maybe<HoverFields>
    }
> = props => {
    const references = useMemo(() => props.references.nodes.map(buildLocation), [props.references])
    const definitions = useMemo(() => props.definitions.nodes.map(buildLocation), [props.definitions])
    const implementations = useMemo(() => props.implementations.nodes.map(buildLocation), [props.implementations])

    return (
        <>
            {props.hover && (
                <Markdown
                    className={classNames('mb-0 card-body text-small', styles.hoverMarkdown)}
                    dangerousInnerHTML={renderMarkdown(props.hover.markdown.text)}
                />
            )}
            <CardHeader>
                <h4 className="p-1 mb-0">Definitions</h4>
            </CardHeader>
            {definitions.length > 0 ? (
                <LocationsList
                    locations={definitions}
                    activeLocation={props.activeLocation}
                    setActiveLocation={props.setActiveLocation}
                    filter={props.filter}
                />
            ) : (
                <p className="text-muted my-1 pl-2">
                    {props.filter ? (
                        <i>
                            No definitions matching <strong>{props.filter}</strong> found
                        </i>
                    ) : (
                        <i>No definitions found</i>
                    )}
                </p>
            )}
            <CardHeader>
                <h4 className="p-1 mb-0">References</h4>
            </CardHeader>
            {references.length > 0 ? (
                <LocationsList
                    locations={references}
                    activeLocation={props.activeLocation}
                    setActiveLocation={props.setActiveLocation}
                    filter={props.filter}
                />
            ) : (
                <p className="text-muted pl-2">
                    {props.filter ? (
                        <i>
                            No references matching <strong>{props.filter}</strong> found
                        </i>
                    ) : (
                        <i>No references found</i>
                    )}
                </p>
            )}
            {implementations.length > 0 && (
                <>
                    <CardHeader>
                        <h4 className="p-1 mb-0">Implementations</h4>
                    </CardHeader>
                    <LocationsList
                        locations={implementations}
                        activeLocation={props.activeLocation}
                        setActiveLocation={props.setActiveLocation}
                        filter={props.filter}
                    />
                </>
            )}
        </>
    )
}

const SideBlob: React.FunctionComponent<
    CoolCodeIntelProps & {
        activeLocation: Location
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
            onTokenClick={(token: CoolClickedToken) => {
                if (props.onTokenClick) {
                    props.onTokenClick(token)
                }
            }}
            coolCodeIntelEnabled={true}
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

const buildFileURL = (location: Location): string => {
    const path = `/${location.resource.repository.name}/-/blob/${location.resource.path}`
    const range = location.range

    if (range !== undefined) {
        return appendSubtreeQueryParameter(
            appendLineRangeQueryParameter(
                path,
                toPositionOrRangeQueryParameter({
                    range: {
                        // ATTENTION: Another off-by-one chaos in the making here
                        start: {
                            line: range.start.line + 1,
                            character: range.start.character + 1,
                        },
                        end: { line: range.end.line + 1, character: range.end.character + 1 },
                    },
                })
            )
        )
    }
    return path
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
                className="bg-transparent border-bottom border-top-0 border-left-0 border-right-0 d-flex justify-content-start w-100"
            >
                {isOpen ? (
                    <MenuUpIcon className={classNames('icon-inline', styles.chevron)} />
                ) : (
                    <MenuDownIcon className={classNames('icon-inline', styles.chevron)} />
                )}

                <span>
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
                className="bg-transparent border-bottom border-top-0 border-left-0 border-right-0 d-flex justify-content-start w-100"
            >
                {isOpen ? (
                    <MenuUpIcon className={classNames('icon-inline', styles.chevron)} />
                ) : (
                    <MenuDownIcon className={classNames('icon-inline', styles.chevron)} />
                )}

                <span className={styles.coolCodeIntelReferenceFilename}>
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

const ResizableCoolCodeIntelPanel = React.memo<CoolCodeIntelProps & { handlePanelClose: (closed: boolean) => void }>(
    props => (
        <Resizable
            className={styles.resizablePanel}
            handlePosition="top"
            defaultSize={350}
            storageKey="panel-size"
            element={<CoolCodeIntelPanel {...props} />}
        />
    )
)

const CoolCodeIntelPanel = React.memo<CoolCodeIntelProps & { handlePanelClose: (closed: boolean) => void }>(props => {
    const [tabIndex, setTabIndex] = useLocalStorage(LAST_TAB_STORAGE_KEY, 0)
    const handleTabsChange = useCallback((index: number) => setTabIndex(index), [setTabIndex])

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
                        onClick={() => props.handlePanelClose(true)}
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

const CoolCodeIntelResizablePanel: React.FunctionComponent<CoolCodeIntelProps & { onClose: () => void }> = props => {
    let token = props.clickedToken

    const history = useHistory()
    const location = useLocation()

    const [closed, close] = useState(false)
    const handlePanelClose = useCallback(() => {
        // Signal up that panel is closed
        props.onClose()
        // Remove 'viewState' from location
        history.push(locationWithoutViewState(location))
        // close(true)
    }, [props, history, location])

    useEffect(() => {
        if (token) {
            close(false)
        }
    }, [token])

    if (closed) {
        return null
    }

    const { hash, pathname, search } = location
    const { line, character, viewState } = parseQueryAndHash(search, hash)

    // If we don't have a token that someone clicked on and we don't have
    // '#tab=...' in the URL, we don't need to show the panel.
    if (!token && !viewState) {
        return null
    }

    const { filePath, repoName, revision, commitID } = parseBrowserRepoURL(pathname)

    const tokenSameAsUrl =
        token &&
        token.repoName === repoName &&
        token.line === line &&
        token.character === character &&
        token.filePath === filePath

    const haveFileLocationAndViewState =
        line &&
        character &&
        filePath &&
        viewState &&
        (viewState === 'references' || viewState.startsWith('implementations_'))

    // If we have info in URL, no clicked token, or the clicked token is not the
    // same as what's in the URL, we use what's in the URL as token
    if (haveFileLocationAndViewState && (!token || !tokenSameAsUrl)) {
        const urlBasedToken = {
            repoName,
            line,
            character,
            filePath,
        }
        if (commitID === undefined || revision === undefined) {
            return <CoolCodeIntelPanelUrlBased {...props} {...urlBasedToken} handlePanelClose={handlePanelClose} />
        }

        token = { ...urlBasedToken, revision, commitID }
    }

    return <ResizableCoolCodeIntelPanel {...props} clickedToken={token} handlePanelClose={handlePanelClose} />
}

export const CoolCodeIntelPanelUrlBased: React.FunctionComponent<
    CoolCodeIntelProps & {
        repoName: string
        line: number
        character: number
        filePath: string
        revision?: string

        handlePanelClose: (closed: boolean) => void
    }
> = props => {
    const resolvedRevision = useObservable(
        useMemo(() => resolveRevision({ repoName: props.repoName, revision: props.revision }), [
            props.repoName,
            props.revision,
        ])
    )

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

    return <ResizableCoolCodeIntelPanel {...props} clickedToken={token} handlePanelClose={props.handlePanelClose} />
}
