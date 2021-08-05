import classNames from 'classnames'
import * as H from 'history'
import React, { useState } from 'react'
import { useLocation } from 'react-router'

import { CircleChevronLeftIcon } from '@sourcegraph/shared/src/components/icons'
import { GitRefType, Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { createAggregateError } from '@sourcegraph/shared/src/util/errors'
import { useConnection } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionForm,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'
import { useDebounce } from '@sourcegraph/wildcard'

import { GitRefFields, RepositoryGitRefsResult, RepositoryGitRefsVariables } from '../graphql-operations'

import { GitReferenceNode, REPOSITORY_GIT_REFS } from './GitReference'

interface GitReferencePopoverNodeProps {
    node: GitRefFields

    defaultBranch: string
    currentRevision: string | undefined

    location: H.Location

    getURLFromRevision: (href: string, revision: string) => string
}

const GitReferencePopoverNode: React.FunctionComponent<GitReferencePopoverNodeProps> = ({
    node,
    defaultBranch,
    currentRevision,
    location,
    getURLFromRevision,
}) => {
    let isCurrent: boolean
    if (currentRevision) {
        isCurrent = node.name === currentRevision || node.abbrevName === currentRevision
    } else {
        isCurrent = node.name === `refs/heads/${defaultBranch}`
    }
    return (
        <GitReferenceNode
            node={node}
            url={getURLFromRevision(location.pathname + location.search + location.hash, node.abbrevName)}
            ancestorIsLink={false}
            className={classNames(
                'connection-popover__node-link',
                isCurrent && 'connection-popover__node-link--active'
            )}
        >
            {isCurrent && (
                <CircleChevronLeftIcon
                    className="icon-inline connection-popover__node-link-icon"
                    data-tooltip="Current"
                />
            )}
        </GitReferenceNode>
    )
}

interface RevisionReferencesTabProps {
    type: GitRefType
    repo: Scalars['ID']
    defaultBranch: string
    getURLFromRevision: (href: string, revision: string) => string

    noun: string
    pluralNoun: string

    /** The current revision, or undefined for the default branch. */
    currentRev: string | undefined

    allowCustomBranches?: boolean
}

export const RevisionReferencesTab: React.FunctionComponent<RevisionReferencesTabProps> = ({
    type,
    repo,
    defaultBranch,
    getURLFromRevision,
    currentRev,
    noun,
    pluralNoun,
}) => {
    const [searchValue, setSearchValue] = useState('')
    const debouncedSearchValue = useDebounce(searchValue, 200)
    const location = useLocation()

    const { connection, loading, errors, hasNextPage, fetchMore } = useConnection<
        RepositoryGitRefsResult,
        RepositoryGitRefsVariables,
        GitRefFields
    >({
        query: REPOSITORY_GIT_REFS,
        variables: {
            query: debouncedSearchValue,
            first: 50,
            repo,
            type,
            withBehindAhead: false,
        },
        getConnection: ({ data, errors }) => {
            if (!data || !data.node || !data.node.gitRefs) {
                throw createAggregateError(errors)
            }
            return data.node.gitRefs
        },
    })

    if (loading) {
        console.log('Loading...')
    }

    if (errors) {
        console.log('Got errors', errors)
    }

    if (connection) {
        console.log('Got connection!')
        console.log(connection.nodes)
    }

    const summary = connection && (
        <ConnectionSummary
            connection={connection}
            noun={noun}
            pluralNoun={pluralNoun}
            totalCount={connection.totalCount ?? null}
            hasNextPage={hasNextPage}
        />
    )

    return (
        <ConnectionContainer compact={true} className="connection-popover__content">
            <ConnectionForm
                inputValue={searchValue}
                onInputChange={event => setSearchValue(event.target.value)}
                autoFocus={true}
                inputPlaceholder="Find..."
                inputClassName="connection-popover__input"
            />
            <SummaryContainer>{searchValue && summary}</SummaryContainer>
            <ConnectionList className="connection-popover__nodes">
                {connection?.nodes?.map((node, index) => (
                    <GitReferencePopoverNode
                        key={index}
                        node={node}
                        currentRevision={currentRev}
                        defaultBranch={defaultBranch}
                        getURLFromRevision={getURLFromRevision}
                        location={location}
                    />
                ))}
            </ConnectionList>
            {loading && <ConnectionLoading />}
            {connection && (
                <SummaryContainer>
                    {summary}
                    {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                </SummaryContainer>
            )}
        </ConnectionContainer>
    )
}
