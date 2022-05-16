import React from 'react'

import ImportIcon from 'mdi-react/ImportIcon'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { Icon, Typography } from '@sourcegraph/wildcard'

import { UseConnectionResult } from '../../../../../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../../../../components/FilteredConnection/ui'

import { ImportingChangesetFields } from './useImportingChangesets'

import styles from './ImportingChangesetsPreviewList.module.scss'

interface ImportingChangesetsPreviewListProps {
    importingChangesetsConnection: UseConnectionResult<ImportingChangesetFields>
    /**
     * Whether or not the changesets in this list are up-to-date with the current batch
     * spec input YAML in the editor.
     */
    isStale: boolean
}

const CHANGESETS_PER_PAGE_COUNT = 100

export const ImportingChangesetsPreviewList: React.FunctionComponent<
    React.PropsWithChildren<ImportingChangesetsPreviewListProps>
> = ({ importingChangesetsConnection: { connection, hasNextPage, fetchMore, loading }, isStale }) => (
    <ConnectionContainer className="w-100">
        <Typography.H4 className="align-self-start w-100 mt-4">Importing changesets</Typography.H4>
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
                            <Icon as={ImportIcon} />{' '}
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
