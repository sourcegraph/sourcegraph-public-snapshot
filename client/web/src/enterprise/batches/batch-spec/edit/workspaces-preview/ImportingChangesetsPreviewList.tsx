import React from 'react'

import { mdiImport } from '@mdi/js'

import { H3, H4, Icon, LinkOrSpan } from '@sourcegraph/wildcard'

import type { UseShowMorePaginationResult } from '../../../../../components/FilteredConnection/hooks/useShowMorePagination'
import {
    ConnectionContainer,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../../../../components/FilteredConnection/ui'
import type { BatchSpecImportingChangesetsResult } from '../../../../../graphql-operations'

import type { ImportingChangesetFields } from './useImportingChangesets'

import styles from './ImportingChangesetsPreviewList.module.scss'

interface ImportingChangesetsPreviewListProps {
    importingChangesetsConnection: UseShowMorePaginationResult<
        BatchSpecImportingChangesetsResult,
        ImportingChangesetFields
    >
    /**
     * Whether or not the changesets in this list are up-to-date with the current batch
     * spec input YAML in the editor.
     */
    isStale: boolean
}

export const ImportingChangesetsPreviewList: React.FunctionComponent<
    React.PropsWithChildren<ImportingChangesetsPreviewListProps>
> = ({ importingChangesetsConnection: { connection, hasNextPage, fetchMore, loading }, isStale }) => (
    <ConnectionContainer className="w-100">
        <H4 as={H3} className="align-self-start w-100 mt-4">
            Importing changesets
        </H4>
        <ConnectionList className="list-group list-group-flush w-100" aria-label="changesets to be imported">
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
                            <Icon aria-hidden={true} svgPath={mdiImport} />{' '}
                            {node.description.__typename === 'ExistingChangesetReference' &&
                                node.description.baseRepository.name}
                        </LinkOrSpan>{' '}
                        #{node.description.__typename === 'ExistingChangesetReference' && node.description.externalID}
                    </li>
                ) : null
            )}
        </ConnectionList>
        {loading && <ConnectionLoading />}
        {connection && (
            <SummaryContainer centered={true}>
                <ConnectionSummary
                    centered={true}
                    noSummaryIfAllNodesVisible={true}
                    connection={connection}
                    noun="imported changeset"
                    pluralNoun="imported changesets"
                    hasNextPage={hasNextPage}
                    emptyElement={null}
                />
                {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
            </SummaryContainer>
        )}
    </ConnectionContainer>
)
