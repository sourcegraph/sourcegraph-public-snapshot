import * as H from 'history'
import ChevronDoubleLeftIcon from 'mdi-react/ChevronDoubleLeftIcon'
import FileTreeIcon from 'mdi-react/FileTreeIcon'
import React, { useCallback } from 'react'
import { Button } from 'reactstrap'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { Resizable } from '@sourcegraph/shared/src/components/Resizable'
import { ResolvedRevisionSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'
import { Collapsible } from '@sourcegraph/web/src/components/Collapsible'

import { RepositoryFields } from '../../graphql-operations'
import { toDocumentationURL } from '../../util/url'

import { DocumentationIndexNode } from './DocumentationIndexNode'
import { GQLDocumentationNode, GQLDocumentationPathInfo, isExcluded, Tag } from './graphql'

interface Props extends Partial<RevisionSpec>, ResolvedRevisionSpec {
    repo: RepositoryFields
    history: H.History
    location: H.Location
    onToggle: (visible: boolean) => void
    node: GQLDocumentationNode
    depth: number
    pagePathID: string
    pathInfo: GQLDocumentationPathInfo
}

const SIZE_STORAGE_KEY = 'repo-docs-sidebar'
const SIDEBAR_KEY = 'repo-docs-sidebar-toggle'
const SIDEBAR_DEFAULT_VISIBILITY = true

export const getSidebarVisibility = (): boolean => {
    try {
        const item = localStorage.getItem(SIDEBAR_KEY)
        return item ? (JSON.parse(item) as boolean) : SIDEBAR_DEFAULT_VISIBILITY
    } catch {
        return SIDEBAR_DEFAULT_VISIBILITY
    }
}

function nonIndexPathIDs(depth: number, pathInfo: GQLDocumentationPathInfo): string[] {
    const paths = []
    if (depth !== 0 && !pathInfo.isIndex) {
        paths.push(pathInfo.pathID)
    }
    depth++
    for (const child of pathInfo.children) {
        paths.push(...nonIndexPathIDs(depth, child))
    }
    return paths
}

const SubpagesList: React.FunctionComponent<Props> = ({ ...props }) => {
    const childPagePathIDs = nonIndexPathIDs(0, props.pathInfo)

    const max = 10
    const firstFew = childPagePathIDs.length > max ? childPagePathIDs.slice(0, max) : childPagePathIDs
    const remaining = childPagePathIDs.length > max ? childPagePathIDs.slice(max) : []

    return (
        <div className="pl-3 pb-3">
            {firstFew.map(pathID => {
                const url = toDocumentationURL({
                    repoName: props.repo.name,
                    revision: props.revision || '',
                    pathID,
                })
                return (
                    <div key={pathID}>
                        <Link id={'index-' + pathID} to={url} className="text-nowrap">
                            {pathID.slice('/'.length)}&#47;
                        </Link>
                    </div>
                )
            })}
            {remaining && (
                <Collapsible
                    title="..."
                    titleAtStart={true}
                    buttonClassName="repository-documentation-sidebar__show-more-button"
                >
                    {remaining.map(pathID => {
                        const url = toDocumentationURL({
                            repoName: props.repo.name,
                            revision: props.revision || '',
                            pathID,
                        })
                        return (
                            <div key={pathID}>
                                <Link id={'index-' + pathID} to={url} className="text-nowrap">
                                    {pathID.slice('/'.length)}&#47;
                                </Link>
                            </div>
                        )
                    })}
                </Collapsible>
            )}
        </div>
    )
}

/**
 * The sidebar for a specific repo revision that shows the index of all documentation.
 */
export const RepositoryDocumentationSidebar: React.FunctionComponent<Props> = ({ onToggle, ...props }) => {
    const [toggleSidebar, setToggleSidebar] = useLocalStorage(SIDEBAR_KEY, SIDEBAR_DEFAULT_VISIBILITY)
    const handleSidebarToggle = useCallback(() => {
        onToggle(!toggleSidebar)
        setToggleSidebar(!toggleSidebar)
    }, [setToggleSidebar, toggleSidebar, onToggle])

    if (!toggleSidebar) {
        return (
            <button
                type="button"
                className="position-absolute btn btn-icon btn-link border-right border-bottom rounded-0 repo-revision-container__toggle"
                onClick={handleSidebarToggle}
                data-tooltip="Show sidebar"
            >
                <FileTreeIcon className="icon-inline" />
            </button>
        )
    }
    const excludingTags: Tag[] = ['private']

    return (
        <Resizable
            defaultSize={384}
            handlePosition="right"
            storageKey={SIZE_STORAGE_KEY}
            element={
                <div className="repository-documentation-sidebar d-flex flex-column w-100 border-right">
                    <div className="d-flex flex-0 mx-3">
                        <Button
                            onClick={handleSidebarToggle}
                            className="bg-transparent border-0 ml-auto p-1 position-relative focus-behaviour"
                            title="Close panel"
                            data-tooltip="Collapse panel"
                            data-placement="right"
                        >
                            <ChevronDoubleLeftIcon className="icon-inline" />
                        </Button>
                    </div>
                    <div aria-hidden={true} className="repository-documentation-sidebar-scroller overflow-auto px-3">
                        {props.pathInfo.isIndex && (
                            <>
                                <h4 className="text-nowrap">Index</h4>
                                {props.pathInfo.children.length > 0 ? (
                                    <SubpagesList onToggle={onToggle} {...props} />
                                ) : (
                                    <p>Looks like there's nothing to see here..</p>
                                )}
                            </>
                        )}
                        {!props.pathInfo.isIndex && props.pathInfo.children.length > 0 && (
                            <>
                                <h4 className="text-nowrap">Subpages</h4>
                                <SubpagesList onToggle={onToggle} {...props} />
                            </>
                        )}
                        {!props.pathInfo.isIndex &&
                            props.pathInfo.children.length === 0 &&
                            isExcluded(props.node, excludingTags) && (
                                <>
                                    <p>Looks like there's nothing to see here..</p>
                                </>
                            )}
                        <DocumentationIndexNode
                            {...props}
                            node={props.node}
                            pagePathID={props.pagePathID}
                            depth={0}
                            contentOnly={false}
                            excludingTags={excludingTags}
                        />
                    </div>
                </div>
            }
        />
    )
}
