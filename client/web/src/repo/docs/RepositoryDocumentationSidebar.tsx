import * as H from 'history'
import ChevronDoubleLeftIcon from 'mdi-react/ChevronDoubleLeftIcon'
import FileTreeIcon from 'mdi-react/FileTreeIcon'
import React, { useCallback } from 'react'
import { Button } from 'reactstrap'

import { Resizable } from '@sourcegraph/shared/src/components/Resizable'
import { ResolvedRevisionSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { RepositoryFields } from '../../graphql-operations'

import { DocumentationIndexNode } from './DocumentationIndexNode'
import { GQLDocumentationNode } from './DocumentationNode'

interface Props extends Partial<RevisionSpec>, ResolvedRevisionSpec {
    repo: RepositoryFields
    className: string
    history: H.History
    location: H.Location
    onToggle: (visible: boolean) => void
    node: GQLDocumentationNode
    depth: number
    pagePathID: string
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

    return (
        <Resizable
            defaultSize={384}
            handlePosition="right"
            storageKey={SIZE_STORAGE_KEY}
            className={props.className}
            element={
                <div className="d-flex flex-column w-100 border-right">
                    <div className="d-flex flex-1 mx-3">
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
                    <div aria-hidden={true} className="d-flex explorer overflow-auto px-3">
                        <DocumentationIndexNode {...props} node={props.node} pagePathID={props.pagePathID} depth={0} />
                    </div>
                </div>
            }
        />
    )
}
