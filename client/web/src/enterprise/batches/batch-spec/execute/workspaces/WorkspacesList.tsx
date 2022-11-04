import React from 'react'

import { useHistory } from 'react-router'

import { UseConnectionResult } from '../../../../../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../../../../components/FilteredConnection/ui'
import {
    HiddenBatchSpecWorkspaceListFields,
    Scalars,
    VisibleBatchSpecWorkspaceListFields,
} from '../../../../../graphql-operations'

import { WorkspacesListItem } from './WorkspacesListItem'

export interface WorkspacesListProps {
    workspacesConnection: UseConnectionResult<VisibleBatchSpecWorkspaceListFields | HiddenBatchSpecWorkspaceListFields>
    /** The currently selected workspace node id. Will be highlighted. */
    selectedNode?: Scalars['ID']
    /** The URL path to the execution page this workspaces list is shown on. */
    executionURL: string
}

export const WorkspacesList: React.FunctionComponent<React.PropsWithChildren<WorkspacesListProps>> = ({
    workspacesConnection: { connection, error, loading, hasNextPage, fetchMore },
    selectedNode,
    executionURL,
}) => {
    const history = useHistory()

    return (
        <ConnectionContainer>
            {error && <ConnectionError errors={[error.message]} />}
            <ConnectionList as="ul" className="mb-0">
                {connection?.nodes?.map(node => (
                    <WorkspacesListItem
                        key={node.id}
                        workspace={node}
                        isSelected={node.id === selectedNode}
                        onSelect={() =>
                            history.replace({
                                ...history.location,
                                pathname: `${executionURL}/execution/workspaces/${node.id}`,
                            })
                        }
                    />
                ))}
            </ConnectionList>
            {/* We don't want to flash a loader on reloads: */}
            {loading && !connection && <ConnectionLoading />}
            {connection && (
                <SummaryContainer centered={true}>
                    <ConnectionSummary
                        centered={true}
                        noSummaryIfAllNodesVisible={true}
                        first={20}
                        connection={connection}
                        noun="workspace"
                        pluralNoun="workspaces"
                        hasNextPage={hasNextPage}
                    />
                    {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                </SummaryContainer>
            )}
        </ConnectionContainer>
    )
}
