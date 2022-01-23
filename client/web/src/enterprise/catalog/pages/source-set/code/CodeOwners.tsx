import React from 'react'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { useConnection } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionLoading,
    SummaryContainer,
    ConnectionSummary,
    ShowMoreButton,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'

import {
    SourceSetCodeOwnersFields,
    SourceSetCodeOwnersResult,
    SourceSetCodeOwnersVariables,
} from '../../../../../graphql-operations'
import { personLinkFieldsFragment } from '../../../../../person/PersonLink'
import { PersonList } from '../../../components/person-list/PersonList'

interface Props {
    sourceSet: Scalars['ID']
    className?: string
}

const SOURCE_SET_CODE_OWNERS = gql`
    query SourceSetCodeOwners($node: ID!, $first: Int!) {
        node(id: $node) {
            ... on SourceSet {
                codeOwners(first: $first) {
                    edges {
                        node {
                            ...PersonLinkFields
                            avatarURL
                        }
                        fileCount
                        fileProportion
                    }
                    totalCount
                    pageInfo {
                        hasNextPage
                    }
                }
            }
        }
    }
    ${personLinkFieldsFragment}
`

const FIRST = 50

export const SourceSetCodeOwners: React.FunctionComponent<Props> = ({ sourceSet, className }) => {
    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        SourceSetCodeOwnersResult,
        SourceSetCodeOwnersVariables,
        NonNullable<SourceSetCodeOwnersFields['codeOwners']>['edges'][number]
    >({
        query: SOURCE_SET_CODE_OWNERS,
        variables: {
            node: sourceSet,
            first: FIRST,
        },
        options: {
            useURL: true,
            fetchPolicy: 'cache-first',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            if (!data.node || !('codeOwners' in data.node) || !data.node.codeOwners) {
                throw new Error('no code owners associated with object')
            }
            return { ...data.node.codeOwners, nodes: data.node.codeOwners.edges }
        },
    })
    return (
        <>
            <h4 className="sr-only">Code owners</h4>
            <ConnectionContainer className={className}>
                {error && <ConnectionError errors={[error.message]} />}
                {connection?.nodes && connection?.nodes.length > 0 && (
                    <PersonList
                        title="Code owners"
                        listTag="ol"
                        orientation="vertical"
                        items={connection.nodes.map(codeOwner => ({
                            person: codeOwner.node,
                            text:
                                codeOwner.fileProportion >= 0.01
                                    ? `${(codeOwner.fileProportion * 100).toFixed(0)}%`
                                    : '<1%',
                            textTooltip: `Owns ${codeOwner.fileCount} ${pluralize('line', codeOwner.fileCount)}`,
                        }))}
                        className={className}
                    />
                )}
                {loading && <ConnectionLoading className="my-2" />}
                {connection && (
                    <SummaryContainer centered={true}>
                        <ConnectionSummary
                            noSummaryIfAllNodesVisible={true}
                            first={FIRST}
                            connection={connection}
                            noun="code owner"
                            pluralNoun="code owners"
                            hasNextPage={hasNextPage}
                            emptyElement={<p>No code owners found</p>}
                        />
                        {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </>
    )
}
