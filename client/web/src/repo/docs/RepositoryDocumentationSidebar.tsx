import classNames from 'classnames'
import * as H from 'history'
import ChevronDoubleLeftIcon from 'mdi-react/ChevronDoubleLeftIcon'
import FileTreeIcon from 'mdi-react/FileTreeIcon'
import React, { useCallback, useMemo } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { Resizable } from '@sourcegraph/shared/src/components/Resizable'
import { ResolvedRevisionSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'
import { Collapsible } from '@sourcegraph/web/src/components/Collapsible'
import { Button } from '@sourcegraph/wildcard'

import { RepositoryFields } from '../../graphql-operations'
import { toDocumentationURL } from '../../util/url'

import { DocumentationIndexNode, IndexNode } from './DocumentationIndexNode'
import { GQLDocumentationNode, GQLDocumentationPathInfo, isExcluded, Tag } from './graphql'
import styles from './RepositoryDocumentationSidebar.module.scss'

interface Props extends Partial<RevisionSpec>, ResolvedRevisionSpec {
    repo: RepositoryFields
    history: H.History
    location: H.Location
    onToggle: (visible: boolean) => void
    node: GQLDocumentationNode
    depth: number
    pagePathID: string

    /** The currently active/visible node's path ID */
    activePathID: string

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
                <Collapsible title="..." titleAtStart={true} buttonClassName={styles.showMoreButton}>
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
export const RepositoryDocumentationSidebar: React.FunctionComponent<Props> = ({
    onToggle,
    node,
    activePathID,
    ...props
}) => {
    const [toggleSidebar, setToggleSidebar] = useLocalStorage(SIDEBAR_KEY, SIDEBAR_DEFAULT_VISIBILITY)
    const handleSidebarToggle = useCallback(() => {
        onToggle(!toggleSidebar)
        setToggleSidebar(!toggleSidebar)
    }, [setToggleSidebar, toggleSidebar, onToggle])

    /**
     * Convert the regular GraphQL node types into IndexNode types. These contain per-node `isActive`
     * and `inActivePath` fields. We bake nodes in this way because otherwise we would need to pass
     * the `activePathID` to every `DocumentationIndexNode` recursively, and it would be an almost-always
     * changing prop to the component - causing scrolling on the page to rerender the entire sidebar
     * instead of just the elements that would've been affected due to `isActive` changes, etc.
     */
    const indexNode = useMemo(() => {
        const bake = (node: GQLDocumentationNode): IndexNode => ({
            ...node,
            children: node.children.map(child =>
                child.pathID ? { pathID: child.pathID } : { node: bake(child.node!) }
            ),
            isActive: node.pathID === activePathID,
            inActivePath: hasDescendent(node, activePathID),
        })
        return bake(node)
    }, [node, activePathID])

    const excludingTags: Tag[] = useMemo(() => ['private'], [])

    if (!toggleSidebar) {
        return (
            <Button
                className="position-absolute btn-icon border-right border-bottom rounded-0 repo-revision-container__toggle"
                onClick={handleSidebarToggle}
                data-tooltip="Show sidebar"
                variant="link"
            >
                <FileTreeIcon className="icon-inline" />
            </Button>
        )
    }

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
                    <div
                        aria-hidden={true}
                        className={classNames('overflow-auto px-3', styles.repositoryDocumentationSidebarScroller)}
                    >
                        {props.pathInfo.isIndex && (
                            <>
                                <h4 className="text-nowrap">Index</h4>
                                {props.pathInfo.children.length > 0 ? (
                                    <SubpagesList
                                        onToggle={onToggle}
                                        {...props}
                                        node={node}
                                        activePathID={activePathID}
                                    />
                                ) : (
                                    <p>Looks like there's nothing to see here..</p>
                                )}
                            </>
                        )}
                        {!props.pathInfo.isIndex && props.pathInfo.children.length > 0 && (
                            <>
                                <h4 className="text-nowrap">Subpages</h4>
                                <SubpagesList onToggle={onToggle} {...props} node={node} activePathID={activePathID} />
                            </>
                        )}
                        {!props.pathInfo.isIndex &&
                            props.pathInfo.children.length === 0 &&
                            isExcluded(node, excludingTags) && (
                                <>
                                    <p>Looks like there's nothing to see here..</p>
                                </>
                            )}
                        <DocumentationIndexNode
                            {...props}
                            node={indexNode}
                            pagePathID={props.pagePathID}
                            depth={0}
                            excludingTags={excludingTags}
                        />
                    </div>
                </div>
            }
        />
    )
}

export function hasDescendent(node: GQLDocumentationNode, descendentPathID: string): boolean {
    return !!node.children.find(child => {
        if (child.pathID === descendentPathID || child.node?.pathID === descendentPathID) {
            return true
        }
        return child.node ? hasDescendent(child.node, descendentPathID) : false
    })
}
