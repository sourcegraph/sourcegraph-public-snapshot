import React, { useState } from 'react'

import type * as H from 'history'
import SearchIcon from 'mdi-react/SearchIcon'
import { useLocation } from 'react-router-dom'

import { createAggregateError, escapeRevspecForURL } from '@sourcegraph/common'
import type { GitRefType, Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { useDebounce } from '@sourcegraph/wildcard'

import { useShowMorePagination } from '../../components/FilteredConnection/hooks/useShowMorePagination'
import { ConnectionSummary } from '../../components/FilteredConnection/ui'
import type { GitRefFields, RepositoryGitRefsResult, RepositoryGitRefsVariables } from '../../graphql-operations'
import { type GitReferenceNodeProps, REPOSITORY_GIT_REFS } from '../GitReference'

import { ConnectionPopoverGitReferenceNode } from './components'
import { RevisionsPopoverTab } from './RevisionsPopoverTab'

interface GitReferencePopoverNodeProps extends Pick<GitReferenceNodeProps, 'node' | 'onClick'> {
    defaultBranch: string
    currentRevision: string | undefined

    location: H.Location

    getPathFromRevision: (href: string, revision: string) => string

    isSpeculative?: boolean

    isPackageVersion?: boolean
}

const GitReferencePopoverNode: React.FunctionComponent<React.PropsWithChildren<GitReferencePopoverNodeProps>> = ({
    node,
    defaultBranch,
    currentRevision,
    location,
    getPathFromRevision,
    isSpeculative,
    onClick,
    isPackageVersion,
}) => {
    let isCurrent: boolean
    if (currentRevision) {
        isCurrent = node.name === currentRevision || node.abbrevName === currentRevision
    } else {
        isCurrent = node.name === `refs/heads/${defaultBranch}`
    }
    return (
        <ConnectionPopoverGitReferenceNode
            node={node}
            url={getPathFromRevision(location.pathname + location.search + location.hash, node.abbrevName)}
            ancestorIsLink={false}
            active={isCurrent}
            onClick={onClick}
            icon={isSpeculative ? SearchIcon : undefined}
            isPackageVersion={isPackageVersion}
        />
    )
}

interface SpectulativeGitReferencePopoverNodeProps
    extends Pick<RevisionsPopoverReferencesProps, 'onSelect'>,
        Omit<GitReferencePopoverNodeProps, 'node'> {
    name: string
    repoName: string
    existingNodes: GitRefFields[]
}

export const SpectulativeGitReferencePopoverNode: React.FunctionComponent<
    React.PropsWithChildren<SpectulativeGitReferencePopoverNodeProps>
> = ({ name, repoName, currentRevision, defaultBranch, getPathFromRevision, location, existingNodes, onSelect }) => {
    const alreadyExists = existingNodes.some(existingNode => existingNode.abbrevName === name)

    if (alreadyExists) {
        // We're already showing this node, so don't show it again.
        return null
    }

    /**
     * A dummy GitReferenceNode that we can use to render a possible result in the same styles as existing, known, results.
     */
    const speculativeGitNode: GitReferenceNodeProps['node'] | null = {
        id: name,
        name,
        displayName: name,
        abbrevName: name,
        url: `/${repoName}@${escapeRevspecForURL(name)}`,
        target: { commit: null },
    }

    // We haven't found a node with the same name, render a node with expected props
    return (
        <GitReferencePopoverNode
            node={speculativeGitNode}
            currentRevision={currentRevision}
            defaultBranch={defaultBranch}
            getPathFromRevision={getPathFromRevision}
            location={location}
            onClick={() => onSelect?.(speculativeGitNode)}
            isSpeculative={true}
        />
    )
}

interface RevisionsPopoverReferencesProps {
    type: GitRefType
    repo: Scalars['ID']
    repoName: string
    defaultBranch: string
    getPathFromRevision: (href: string, revision: string) => string

    noun: string
    pluralNoun: string

    /** The current revision, or undefined for the default branch. */
    currentRev: string | undefined

    showSpeculativeResults?: boolean

    isPackage?: boolean

    onSelect?: (node: GitRefFields) => void

    tabLabel: string
}

const BATCH_COUNT = 50

export const RevisionsPopoverReferences: React.FunctionComponent<
    React.PropsWithChildren<RevisionsPopoverReferencesProps>
> = ({
    type,
    repo,
    repoName,
    defaultBranch,
    getPathFromRevision,
    currentRev,
    noun,
    pluralNoun,
    showSpeculativeResults,
    onSelect,
    tabLabel,
    isPackage,
}) => {
    const [searchValue, setSearchValue] = useState('')
    const query = useDebounce(searchValue, 200)
    const location = useLocation()

    const response = useShowMorePagination<RepositoryGitRefsResult, RepositoryGitRefsVariables, GitRefFields>({
        query: REPOSITORY_GIT_REFS,
        variables: {
            query,
            first: BATCH_COUNT,
            repo,
            type,
            withBehindAhead: false,
        },
        getConnection: ({ data, errors }) => {
            if (data?.node?.__typename !== 'Repository' || !data?.node?.gitRefs) {
                throw createAggregateError(errors)
            }
            return data.node.gitRefs
        },
        options: {
            fetchPolicy: 'cache-first',
        },
    })

    const summary = response.connection && (
        <ConnectionSummary
            emptyElement={showSpeculativeResults ? <></> : undefined}
            connection={response.connection}
            first={BATCH_COUNT}
            noun={noun}
            pluralNoun={pluralNoun}
            hasNextPage={response.hasNextPage}
            connectionQuery={query}
            compact={true}
        />
    )

    return (
        <RevisionsPopoverTab
            {...response}
            query={query}
            summary={summary}
            inputValue={searchValue}
            onInputChange={setSearchValue}
            inputAriaLabel={tabLabel}
        >
            {response.connection?.nodes.map((node, index) => (
                <GitReferencePopoverNode
                    key={index}
                    node={node}
                    currentRevision={currentRev}
                    defaultBranch={defaultBranch}
                    getPathFromRevision={getPathFromRevision}
                    location={location}
                    onClick={() => onSelect?.(node)}
                    isPackageVersion={isPackage}
                />
            ))}
            {showSpeculativeResults && response.connection && query && (
                <SpectulativeGitReferencePopoverNode
                    name={query}
                    repoName={repoName}
                    existingNodes={response.connection.nodes}
                    currentRevision={currentRev}
                    defaultBranch={defaultBranch}
                    getPathFromRevision={getPathFromRevision}
                    location={location}
                    onSelect={onSelect}
                />
            )}
        </RevisionsPopoverTab>
    )
}
