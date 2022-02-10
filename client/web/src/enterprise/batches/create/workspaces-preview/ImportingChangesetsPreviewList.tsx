import ImportIcon from 'mdi-react/ImportIcon'
import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { UseConnectionResult } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'

import styles from './ImportingChangesetsPreviewList.module.scss'
import { ImportingChangesetFields } from './useImportingChangesets'

interface ImportingChangesetsPreviewListProps {
    importingChangesetsConnection: UseConnectionResult<ImportingChangesetFields>
    /**
     * Whether or not the changesets in this list are up-to-date with the current batch
     * spec input YAML in the editor.
     */
    isStale: boolean
}

const CHANGESETS_PER_PAGE_COUNT = 100

export const ImportingChangesetsPreviewList: React.FunctionComponent<ImportingChangesetsPreviewListProps> = ({
    importingChangesetsConnection: { connection, hasNextPage, fetchMore, loading },
    isStale,
}) => (
    <ConnectionContainer className="w-100">
        <h4 className="align-self-start w-100 mt-4">Importing changesets</h4>
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
                        #{node.description.__typename === 'ExistingChangesetReference' && node.description.externalID}
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
