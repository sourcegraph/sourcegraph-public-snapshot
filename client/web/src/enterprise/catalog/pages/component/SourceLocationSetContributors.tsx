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
    SourceLocationSetContributorsFields,
    SourceLocationSetContributorsResult,
    SourceLocationSetContributorsVariables,
} from '../../../../graphql-operations'
import { personLinkFieldsFragment } from '../../../../person/PersonLink'

import { PersonList } from '../../components/person-list/PersonList'

interface Props {
    sourceLocationSet: Scalars['ID']
    className?: string
}

const SOURCE_LOCATION_SET_CONTRIBUTORS = gql`
    query SourceLocationSetContributors($node: ID!, $first: Int!) {
        node(id: $node) {
            ... on SourceLocationSet {
                contributors(first: $first) {
                    edges {
                        person {
                            ...PersonLinkFields
                            avatarURL
                        }
                        authoredLineCount
                        authoredLineProportion
                        lastCommit {
                            author {
                                date
                            }
                        }
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

export const SourceLocationSetContributors: React.FunctionComponent<Props> = ({ sourceLocationSet, className }) => {
    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        SourceLocationSetContributorsResult,
        SourceLocationSetContributorsVariables,
        NonNullable<SourceLocationSetContributorsFields['contributors']>['edges'][number]
    >({
        query: SOURCE_LOCATION_SET_CONTRIBUTORS,
        variables: {
            node: sourceLocationSet,
            first: FIRST,
        },
        options: {
            useURL: true,
            fetchPolicy: 'cache-first',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            if (!data.node || !('contributors' in data.node) || !data.node.contributors) {
                throw new Error('no code owners associated with object')
            }
            return { ...data.node.contributors, nodes: data.node.contributors.edges }
        },
    })
    return (
        <>
            <h4 className="sr-only">Contributors</h4>
            <ConnectionContainer className={className}>
                {error && <ConnectionError errors={[error.message]} />}
                {connection?.nodes && connection?.nodes.length > 0 && (
                    <PersonList
                        title="Contributors"
                        listTag="ol"
                        orientation="vertical"
                        items={connection.nodes.map(contributor => ({
                            person: contributor.person,
                            text:
                                contributor.authoredLineProportion >= 0.01
                                    ? `${(contributor.authoredLineProportion * 100).toFixed(0)}%`
                                    : '<1%',
                            textTooltip: `${contributor.authoredLineCount} ${pluralize(
                                'line',
                                contributor.authoredLineCount
                            )}`,
                            date: contributor.lastCommit.author.date,
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
                            noun="contributor"
                            pluralNoun="contributors"
                            hasNextPage={hasNextPage}
                            emptyElement={<p>No contributors found</p>}
                        />
                        {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </>
    )
}
