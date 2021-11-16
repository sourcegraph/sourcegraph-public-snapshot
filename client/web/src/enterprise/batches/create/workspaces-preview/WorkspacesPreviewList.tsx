import ImportIcon from 'mdi-react/ImportIcon'
import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'

import {
    WorkspacesAndImportingChangesetsResult,
    WorkspacesAndImportingChangesetsVariables,
    Scalars,
} from '../../../../graphql-operations'
import { WORKSPACES_AND_IMPORTING_CHANGESETS } from '../backend'

import { PreviewLoadingSpinner } from './PreviewLoadingSpinner'
import { WorkspacesPreviewListItem } from './WorkspacesPreviewListItem'

interface WorkspacesPreviewListProps {
    batchSpecID: Scalars['ID']
    /**
     * Whether or not the workspaces in this list are up-to-date with the current batch
     * spec input YAML in the editor.
     */
    isStale: boolean
    setResolutionError: (error: string) => void
    /**
     * Function to automatically update repo query of input batch spec YAML to exclude the
     * provided repo + branch.
     */
    excludeRepo: (repo: string, branch: string) => void
}

export const WorkspacesPreviewList: React.FunctionComponent<WorkspacesPreviewListProps> = ({
    batchSpecID,
    isStale,
    setResolutionError,
    excludeRepo,
}) => {
    const { data, loading } = useQuery<
        WorkspacesAndImportingChangesetsResult,
        WorkspacesAndImportingChangesetsVariables
    >(WORKSPACES_AND_IMPORTING_CHANGESETS, {
        variables: { batchSpec: batchSpecID },
        // This data is intentionally transient, so there's no need to cache it.
        fetchPolicy: 'no-cache',
        // Report Apollo client errors back to the parent.
        onError: error => setResolutionError(error.message),
    })

    if (loading) {
        return <PreviewLoadingSpinner className="my-4" />
    }

    const workspaces = data?.node?.__typename === 'BatchSpec' ? data.node.workspaceResolution?.workspaces : undefined
    const importingChangesets = data?.node?.__typename === 'BatchSpec' ? data.node.importingChangesets : undefined

    return (
        <>
            {!workspaces || workspaces.nodes.length === 0 ? (
                <span className="text-muted">No workspaces found</span>
            ) : (
                <ul className="list-group p-1 mb-0 w-100">
                    {workspaces?.nodes.map((item, index) => (
                        <WorkspacesPreviewListItem
                            key={`${item.repository.id}-${item.branch.id}`}
                            item={item}
                            isStale={isStale}
                            exclude={excludeRepo}
                            variant={index % 2 === 0 ? 'light' : 'dark'}
                        />
                    ))}
                </ul>
            )}
            {importingChangesets && importingChangesets.totalCount > 0 && (
                <>
                    <h4 className="align-self-start w-100 mt-4">Importing changesets</h4>
                    <ul className="w-100">
                        {importingChangesets?.nodes.map(node =>
                            node.__typename === 'VisibleChangesetSpec' ? (
                                <li className="w-100" key={node.id}>
                                    <LinkOrSpan
                                        to={
                                            node.description.__typename === 'ExistingChangesetReference'
                                                ? node.description.baseRepository.url
                                                : undefined
                                        }
                                    >
                                        <ImportIcon className="icon-inline" />{' '}
                                        {node.description.__typename === 'ExistingChangesetReference' &&
                                            node.description.baseRepository.name}
                                    </LinkOrSpan>{' '}
                                    #
                                    {node.description.__typename === 'ExistingChangesetReference' &&
                                        node.description.externalID}
                                </li>
                            ) : null
                        )}
                    </ul>
                </>
            )}
        </>
    )
}
