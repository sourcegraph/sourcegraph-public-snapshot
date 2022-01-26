import ImportIcon from 'mdi-react/ImportIcon'
import React from 'react'

import { dataOrThrowErrors } from '@sourcegraph/http-client'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import {
    useConnection,
    UseConnectionResult,
} from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'

import {
    Scalars,
    PreviewBatchSpecImportingChangesetFields,
    BatchSpecImportingChangesetsResult,
    BatchSpecImportingChangesetsVariables,
} from '../../../../graphql-operations'
import { IMPORTING_CHANGESETS } from '../backend'

import styles from './ImportingChangesetsPreviewList.module.scss'

interface ImportingChangesetsPreviewListProps {
    batchSpecID: Scalars['ID']
    /**
     * Whether or not the changesets in this list are up-to-date with the current batch
     * spec input YAML in the editor.
     */
    isStale: boolean
}

const CHANGESETS_PER_PAGE_COUNT = 100

export const ImportingChangesetsPreviewList: React.FunctionComponent<ImportingChangesetsPreviewListProps> = ({
    batchSpecID,
    isStale,
}) => {
    const { connection, error, loading, hasNextPage, fetchMore } = useImportingChangesets(batchSpecID)

    if (loading || connection?.totalCount === 0) {
        return null
    }

    return (
        <ConnectionContainer className="w-100">
            <h4 className="align-self-start w-100 mt-4">Importing changesets</h4>
            {error && <ConnectionError errors={[error.message]} />}
            <ConnectionList className="list-group list-group-flush w-100">
                {connection?.nodes.map(node =>
                    node.__typename === 'VisibleChangesetSpec' ? (
                        <li className="w-100" key={node.id}>
                            <LinkOrSpan
                                className={isStale ? styles.stale : undefined}
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
            </ConnectionList>
            {loading && <ConnectionLoading />}
            {connection && (
                <SummaryContainer centered={true}>
                    <ConnectionSummary
                        noSummaryIfAllNodesVisible={true}
                        first={CHANGESETS_PER_PAGE_COUNT}
                        connection={connection}
                        noun="imported changeset"
                        pluralNoun="imported changesets"
                        hasNextPage={hasNextPage}
                        emptyElement={null}
                    />
                    {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                </SummaryContainer>
            )}
        </ConnectionContainer>
    )
}

const useImportingChangesets = (
    batchSpecID: Scalars['ID']
): UseConnectionResult<
    PreviewBatchSpecImportingChangesetFields | { __typename: 'HiddenChangesetSpec'; id: Scalars['ID'] }
> =>
    useConnection<
        BatchSpecImportingChangesetsResult,
        BatchSpecImportingChangesetsVariables,
        PreviewBatchSpecImportingChangesetFields | { __typename: 'HiddenChangesetSpec'; id: Scalars['ID'] }
    >({
        query: IMPORTING_CHANGESETS,
        variables: {
            batchSpec: batchSpecID,
            after: null,
            first: CHANGESETS_PER_PAGE_COUNT,
        },
        options: {
            useURL: false,
            fetchPolicy: 'cache-and-network',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            if (!data.node) {
                throw new Error(`Batch spec with ID ${batchSpecID} does not exist`)
            }
            if (data.node.__typename !== 'BatchSpec') {
                throw new Error(`The given ID is a ${data.node.__typename as string}, not a BatchSpec`)
            }
            if (!data.node.importingChangesets) {
                throw new Error(`No importing changesets resolution found for batch spec with ID ${batchSpecID}`)
            }
            return data.node.importingChangesets
        },
    })
