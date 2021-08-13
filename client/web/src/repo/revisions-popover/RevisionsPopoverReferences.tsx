import classNames from 'classnames'
import * as H from 'history'
import SearchIcon from 'mdi-react/SearchIcon'
import React, { useState } from 'react'
import { useLocation } from 'react-router'

import { GitRefType, Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { createAggregateError } from '@sourcegraph/shared/src/util/errors'
import { useConnection } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import { ConnectionSummary } from '@sourcegraph/web/src/components/FilteredConnection/ui'
import { useDebounce } from '@sourcegraph/wildcard'

import { GitRefFields, RepositoryGitRefsResult, RepositoryGitRefsVariables } from '../../graphql-operations'
import { GitReferenceNode, REPOSITORY_GIT_REFS } from '../GitReference'

import { RevisionsPopoverTab } from './RevisionsPopoverTab'

interface GitReferencePopoverNodeProps {
    node: GitRefFields

    defaultBranch: string
    currentRevision: string | undefined

    location: H.Location

    getURLFromRevision: (href: string, revision: string) => string

    isSpeculative?: boolean
}

const GitReferencePopoverNode: React.FunctionComponent<GitReferencePopoverNodeProps> = ({
    node,
    defaultBranch,
    currentRevision,
    location,
    getURLFromRevision,
    isSpeculative,
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
            icon={isSpeculative ? SearchIcon : undefined}
        />
    )
}

interface SpectulativeGitReferencePopoverNodeProps extends Omit<GitReferencePopoverNodeProps, 'node'> {
    name: string
    repoName: string
    existingNodes: GitRefFields[]
}

export const SpectulativeGitReferencePopoverNode: React.FunctionComponent<SpectulativeGitReferencePopoverNodeProps> = ({
    name,
    repoName,
    currentRevision,
    defaultBranch,
    getURLFromRevision,
    location,
    existingNodes,
}) => {
    const alreadyExists = existingNodes.some(existingNode => existingNode.abbrevName === name)

    if (alreadyExists) {
        // We're already showing this node, so don't show it again.
        return null
    }

    // TODO: Check the node props and ensure URL is handled correctly
    // We haven't found a node with the same name, render a node with expected props
    return (
        <GitReferencePopoverNode
            node={{
                displayName: name,
                name,
                abbrevName: name,
                id: name,
                url: `/${repoName}@${name}`,
                target: { commit: null },
            }}
            currentRevision={currentRevision}
            defaultBranch={defaultBranch}
            getURLFromRevision={getURLFromRevision}
            location={location}
            isSpeculative={true}
        />
    )
}

interface RevisionReferencesTabProps {
    type: GitRefType
    repo: Scalars['ID']
    repoName: string
    defaultBranch: string
    getURLFromRevision: (href: string, revision: string) => string

    noun: string
    pluralNoun: string

    /** The current revision, or undefined for the default branch. */
    currentRev: string | undefined

    allowSpeculativeSearch?: boolean
}

const BATCH_COUNT = 50

export const RevisionReferencesTab: React.FunctionComponent<RevisionReferencesTabProps> = ({
    type,
    repo,
    repoName,
    defaultBranch,
    getURLFromRevision,
    currentRev,
    noun,
    pluralNoun,
    allowSpeculativeSearch,
}) => {
    const [searchValue, setSearchValue] = useState('')
    const query = useDebounce(searchValue, 200)
    const location = useLocation()

    const response = useConnection<RepositoryGitRefsResult, RepositoryGitRefsVariables, GitRefFields>({
        query: REPOSITORY_GIT_REFS,
        variables: {
            query,
            first: BATCH_COUNT,
            repo,
            type,
            withBehindAhead: false,
        },
        getConnection: ({ data, errors }) => {
            if (!data || !data.node || data.node.__typename !== 'Repository' || !data.node.gitRefs) {
                throw createAggregateError(errors)
            }
            return data.node.gitRefs
        },
    })

    const summary = response.connection && (
        <ConnectionSummary
            emptyElement={allowSpeculativeSearch ? <></> : undefined}
            connection={response.connection}
            first={BATCH_COUNT}
            noun={noun}
            pluralNoun={pluralNoun}
            hasNextPage={response.hasNextPage}
            connectionQuery={query}
        />
    )

    return (
        <RevisionsPopoverTab
            {...response}
            query={query}
            summary={summary}
            inputValue={searchValue}
            onInputChange={setSearchValue}
        >
            {response.connection?.nodes?.map((node, index) => (
                <GitReferencePopoverNode
                    key={index}
                    node={node}
                    currentRevision={currentRev}
                    defaultBranch={defaultBranch}
                    getURLFromRevision={getURLFromRevision}
                    location={location}
                />
            ))}
            {/* For branch filtering, we support speculative searching */}
            {allowSpeculativeSearch && response.connection && query && (
                <SpectulativeGitReferencePopoverNode
                    name={query}
                    repoName={repoName}
                    existingNodes={response.connection.nodes}
                    currentRevision={currentRev}
                    defaultBranch={defaultBranch}
                    getURLFromRevision={getURLFromRevision}
                    location={location}
                />
            )}
        </RevisionsPopoverTab>
    )
}
