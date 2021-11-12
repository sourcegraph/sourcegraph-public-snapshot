import CloseIcon from 'mdi-react/CloseIcon'
import ContentSaveIcon from 'mdi-react/ContentSaveIcon'
import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import {
    WorkspacesAndImportingChangesetsResult,
    WorkspacesAndImportingChangesetsVariables,
    Scalars,
} from '../../../../graphql-operations'
import { WORKSPACES_AND_IMPORTING_CHANGESETS } from '../backend'

interface WorkspacesPreviewListProps {
    batchSpecID: Scalars['ID']
    setResolutionError: (error: string) => void
    /**
     * Function to automatically update repo query of input batch spec YAML to exclude the
     * provided repo + branch.
     */
    excludeRepo: (repo: string, branch: string) => void
}

export const WorkspacesPreviewList: React.FunctionComponent<WorkspacesPreviewListProps> = ({
    batchSpecID,
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
        return <LoadingSpinner className="my-4" />
    }

    const workspaces = data?.node?.__typename === 'BatchSpec' ? data.node.workspaceResolution?.workspaces : undefined
    const importingChangesets = data?.node?.__typename === 'BatchSpec' ? data.node.importingChangesets : undefined

    return (
        <>
            {!workspaces || workspaces.nodes.length === 0 ? (
                <span className="text-muted">No workspaces found</span>
            ) : (
                <ul className="list-group p-1 mb-0 w-100">
                    {workspaces?.nodes.map(item => (
                        <li
                            className="d-flex border-bottom mb-3 w-100"
                            key={`${item.repository.id}_${item.branch.target.oid}_${item.path || '/'}`}
                        >
                            <button
                                className="btn align-self-start p-0 m-0 mr-3"
                                data-tooltip="Omit this repository from batch spec file"
                                type="button"
                                // TODO: Show warning that for monorepos, we will currently exclude all paths.
                                onClick={() => excludeRepo(item.repository.name, item.branch.displayName)}
                            >
                                <CloseIcon className="icon-inline" />
                            </button>
                            {item.cachedResultFound && <ContentSaveIcon className="icon-inline" />}
                            <div className="mb-2 flex-1">
                                <p>
                                    {item.repository.name}:{item.branch.abbrevName} Path: {item.path || '/'}
                                </p>
                                <p>
                                    {item.searchResultPaths.length} {pluralize('result', item.searchResultPaths.length)}
                                </p>
                            </div>
                        </li>
                    ))}
                </ul>
            )}
            {importingChangesets && importingChangesets.totalCount > 0 && (
                <>
                    <h4 className="align-self-start w-100">Importing changesets</h4>
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
