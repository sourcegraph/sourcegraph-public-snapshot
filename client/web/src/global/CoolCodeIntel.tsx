import classNames from 'classnames'
import * as H from 'history'
import CloseIcon from 'mdi-react/CloseIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import OpenInAppIcon from 'mdi-react/OpenInAppIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { MemoryRouter, useHistory, useLocation } from 'react-router'
import { Collapse } from 'reactstrap'

import { HoveredToken } from '@sourcegraph/codeintellify'
import { Range } from '@sourcegraph/extension-api-types'
import { useQuery } from '@sourcegraph/http-client'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { Resizable } from '@sourcegraph/shared/src/components/Resizable'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import {
    RepoSpec,
    RevisionSpec,
    FileSpec,
    ResolvedRevisionSpec,
    toPositionOrRangeQueryParameter,
    appendLineRangeQueryParameter,
    appendSubtreeQueryParameter,
    parseQueryAndHash,
    formatSearchParameters,
    addLineRangeQueryParameter,
    lprToRange,
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

import {
    CoolCodeIntelHighlightedBlobResult,
    CoolCodeIntelHighlightedBlobVariables,
    CoolCodeIntelReferencesResult,
    CoolCodeIntelReferencesVariables,
    LocationFields,
    Maybe,
} from '../graphql-operations'
import { resolveRevision } from '../repo/backend'
import { Blob, BlobProps } from '../repo/blob/Blob'
import { parseBrowserRepoURL } from '../util/url'

import styles from './CoolCodeIntel.module.scss'
import { FETCH_HIGHLIGHTED_BLOB, FETCH_REFERENCES_QUERY } from './CoolCodeIntelQueries'

export interface GlobalCoolCodeIntelProps {
    onTokenClick?: (clickedToken: CoolClickedToken) => void
}

type CoolClickedToken = HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec

interface CoolCodeIntelProps extends Omit<BlobProps, 'className' | 'wrapCode' | 'blobInfo' | 'disableStatusBar'> {
    /**
     * The panel runs inside its own MemoryRouter, we keep track of externalHistory
     * so that we're still able to actually navigate within the browser when required
     */
    externalHistory: H.History
    externalLocation: H.Location
}

export const CoolCodeIntel: React.FunctionComponent<CoolCodeIntelProps> = ({ globalHistory, ...props }) => {
    const location = useLocation()
    const { viewState } = parseQueryAndHash(location.search, location.hash)

    // If we don't have '#tab=...' in the URL, we don't need to show the panel.
    if (!viewState) {
        return null
    }

    return (
        <MemoryRouter initialEntries={[location]}>
            <CoolCodeIntelResizablePanel {...props} globalHistory={globalHistory} />
        </MemoryRouter>
    )
}

const LAST_TAB_STORAGE_KEY = 'CoolCodeIntel.lastTab'

type CoolCodeIntelTabID = 'references' | 'token' | 'definition'

interface CoolCodeIntelTab {
    id: CoolCodeIntelTabID
    label: string
    component: React.ComponentType<ReferencesPanelProps>
}

interface CoolCodePanelTabProps extends CoolCodeIntelProps {
    clickedToken: CoolClickedToken
}

interface ReferencesPanelProps extends CoolCodeIntelProps {
    clickedToken?: CoolClickedToken
}

export const ReferencesPanel: React.FunctionComponent<ReferencesPanelProps> = ({ clickedToken, ...props }) => {
    if (!clickedToken) {
        return null
    }

    return <ReferencesList clickedToken={clickedToken} {...props} />
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

interface ReferencesListProps extends CoolCodePanelTabProps {}

export const ReferencesList: React.FunctionComponent<ReferencesListProps> = props => {
    const [filter, setFilter] = useState<string | undefined>(undefined)
    const debouncedFilter = useDebounce(filter, 150)
    const location = useLocation()

    useEffect(() => {
        setFilter(undefined)
    }, [props.clickedToken])

    const matchesExternalURL =
        `${props.externalLocation.pathname}${props.externalLocation.search}` ===
        `${location.pathname}${location.search}`

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
                    <SideReferences {...props} filter={debouncedFilter} />
                </div>
                {props.clickedToken !== undefined && (
                    <div className={classNames('px-0 border-left', styles.referencesSideBlob)}>
                        <SideBlob {...props} />
                    </div>
                )}
            </div>
        </>
    )
}

interface SideReferencesProps extends ReferencesListProps {
    filter: string | undefined
}

export const SideReferences: React.FunctionComponent<SideReferencesProps> = props => {
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

interface LSIFLocationResult {
    __typename?: 'LocationConnection'
    nodes: ({ __typename?: 'Location' } & LocationFields)[]
    pageInfo: { __typename?: 'PageInfo'; endCursor: Maybe<string> }
}

interface SideReferencesListsProps extends SideReferencesProps {
    references: LSIFLocationResult
    definitions: Omit<LSIFLocationResult, 'pageInfo'>
    implementations: LSIFLocationResult
    hover: Maybe<{
        __typename?: 'Hover'
        markdown: { __typename?: 'Markdown'; html: string; text: string }
    }>
}

export const SideReferencesLists: React.FunctionComponent<SideReferencesListsProps> = props => {
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
                    <LocationsList locations={implementations} filter={props.filter} />
                </>
            )}
        </>
    )
}

interface SideBlobProps extends CoolCodePanelTabProps {}

export const SideBlob: React.FunctionComponent<SideBlobProps> = props => {
    const { data, error, loading } = useQuery<
        CoolCodeIntelHighlightedBlobResult,
        CoolCodeIntelHighlightedBlobVariables
    >(FETCH_HIGHLIGHTED_BLOB, {
        variables: {
            repository: props.clickedToken.repoName,
            commit: props.clickedToken.commitID,
            path: props.clickedToken.filePath,
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
                        Loading <code>{props.clickedToken.filePath}</code>...
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

    const {
        content,
        url,
        highlight: { html, aborted },
    } = data?.repository?.commit?.blob

    if (aborted) {
        return (
            <p className="text-warning text-center">
                <i>
                    Highlighting <code>{props.clickedToken.filePath}</code> failed
                </i>
            </p>
        )
    }

    return (
<<<<<<< HEAD
        <Blob
            {...props}
            onTokenClick={(token: CoolClickedToken) => {
                console.log('Called with', token)
                console.log(props.onTokenClick)
                if (props.onTokenClick) {
                    console.log('calling onTokenClick!!')
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
=======
        <>
            <CardHeader className={classNames('pl-1', styles.referencesSideBlobFilename)}>
                <h4>
                    {props.clickedToken.filePath}{' '}
                    <Link to={url}>
                        <OpenInAppIcon className="icon-inline" />
                    </Link>
                </h4>
            </CardHeader>
            <Blob
                {...props}
                disableStatusBar={true}
                wrapCode={true}
                className={styles.referencesSideBlobCode}
                blobInfo={{
                    content,
                    html,
                    filePath: props.clickedToken.filePath,
                    repoName: props.clickedToken.repoName,
                    commitID: props.clickedToken.commitID,
                    revision: props.clickedToken.revision,
                    mode: 'lspmode',
                }}
            />
        </>
>>>>>>> d0166ba981 (Run everything through a memory router)
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
    filter: string | undefined
}> = ({ locations, filter }) => {
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
                    getLineContent={getLineContent}
                    filter={filter}
                />
            ))}
        </>
    )
}

const RepoReferenceGroup: React.FunctionComponent<{
    repoReferenceGroup: RepoLocationGroup
    getLineContent: (location: Location) => string
    filter: string | undefined
}> = ({ repoReferenceGroup, getLineContent, filter }) => {
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
    getLineContent: (reference: Location) => string
    filter: string | undefined
}> = ({ group, getLineContent, filter }) => {
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
                        // TOOD: Fix comparisons from reference to clickedToken
                        const className = reference.url === 'fixme' ? styles.coolCodeIntelReferenceActive : ''

                        return (
                            <li key={reference.url} className={classNames('border-0 rounded-0', className)}>
                                <div>
                                    <Link to={reference.url} className={styles.referenceLink}>
                                        <span>
                                            {reference.range?.start?.line}
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

interface ResizableCoolCodeIntelPanelProps extends CoolCodeIntelPanelProps, CoolCodePanelTabProps {}

export const ResizableCoolCodeIntelPanel = React.memo<ResizableCoolCodeIntelPanelProps>(props => (
    <Resizable
        className={styles.resizablePanel}
        handlePosition="top"
        defaultSize={350}
        storageKey="panel-size"
        element={<CoolCodeIntelPanel {...props} />}
    />
))

interface CoolCodeIntelPanelProps extends CoolCodeIntelProps {
    handlePanelClose: (closed: boolean) => void
}

export const CoolCodeIntelPanel = React.memo<CoolCodeIntelPanelProps>(props => {
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

export const CoolCodeIntelResizablePanel: React.FunctionComponent<CoolCodeIntelProps> = props => {
    const location = useLocation()
    const history = useHistory()

    // TODO: Implement some UI that tracks historical references similar to cs.google
    console.log('Current reference stack:')
    console.log(JSON.stringify(history.entries, null, 2))

    // Experimental reference panel
    const [token, setToken] = useState<CoolClickedToken>()

    const [closed, close] = useState(false)
    const handlePanelClose = useCallback(() => {
        // Signal up that panel is closed
        setToken(undefined)
        // Remove 'viewState' from location
        props.globalHistory.push(locationWithoutViewState(location))
        // close(true)
    }, [props.globalHistory, location])

    useEffect(() => {
        if (token) {
            close(false)
        }
    }, [token])

    if (closed) {
        return null
    }

    const { hash, pathname, search } = location
    const { line, character } = parseQueryAndHash(search, hash)

    const { filePath, repoName } = parseBrowserRepoURL(pathname)

    const haveFileLocation = line && character && filePath

    if (!haveFileLocation) {
        return null
    }

    const urlBasedToken = {
        repoName,
        line,
        character,
        filePath,
    }

    return (
        <CoolCodeIntelPanelUrlBased
            {...props}
            {...urlBasedToken}
            handlePanelClose={handlePanelClose}
            onTokenClick={setToken}
        />
    )
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
