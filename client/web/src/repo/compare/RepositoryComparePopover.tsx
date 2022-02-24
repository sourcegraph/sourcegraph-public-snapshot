import React, { useState } from 'react'
import { Popover } from 'reactstrap'

import { escapeRevspecForURL } from '@sourcegraph/common'
import { Button, Icon } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'
import { RepoRevisionChevronDownIcon } from '../components/RepoRevision'
import { RevisionsPopover } from '../RevisionsPopover'

import { RepositoryCompareAreaPageProps } from './RepositoryCompareArea'

interface RepositoryCompareHeaderProps extends RepositoryCompareAreaPageProps {
    className: string
}

interface RevisionComparison {
    base: RepositoryCompareHeaderProps['base']
    head: RepositoryCompareHeaderProps['head']
}

interface RepositoryComparePopoverProps {
    /**
     * Uniquely identify this specific popover. Used to link the trigger button with the popover to display
     */
    id: string
    /**
     * Initial revision comparison to load. This can be changed by selecting a new revision through the popover.
     */
    comparison: RevisionComparison
    /**
     * The specific comparison type that the popover is concerned with changing.
     */
    type: keyof RevisionComparison
    repo: RepositoryCompareHeaderProps['repo']
}

export const RepositoryComparePopover: React.FunctionComponent<RepositoryComparePopoverProps> = ({
    id,
    comparison,
    repo,
    type,
}) => {
    const [popoverOpen, setPopoverOpen] = useState(false)
    const togglePopover = (): void => setPopoverOpen(previous => !previous)

    const handleSelect = (): void => {
        eventLogger.log('RepositoryComparisonSubmitted')
        togglePopover()
    }

    /**
     * Override the default node URL behavior to support navigating to a repository sub-page.
     */
    const getPathFromRevision = (_href: string, revision: string): string => {
        const escapedRevision = escapeRevspecForURL(revision)
        const comparePath =
            type === 'base'
                ? `${escapedRevision}...${escapeRevspecForURL(comparison.head.revision || '')}`
                : `${escapeRevspecForURL(comparison.base.revision || '')}...${escapedRevision}`

        return `/${repo.name}/-/compare/${comparePath}`
    }

    const defaultBranch = repo.defaultBranch?.abbrevName || 'HEAD'
    const currentRevision = comparison[type]?.revision || undefined

    return (
        <Button
            type="button"
            variant="secondary"
            outline={true}
            className="d-flex align-items-center text-nowrap"
            id={id}
            aria-label={`Change ${type} Git revspec for comparison`}
        >
            <div className="text-muted mr-1">{type}: </div>
            {comparison[type].revision || defaultBranch}
            <Icon as={RepoRevisionChevronDownIcon} />
            <Popover
                isOpen={popoverOpen}
                toggle={togglePopover}
                placement="bottom-start"
                target={id}
                trigger="legacy"
                hideArrow={true}
                fade={false}
                popperClassName="border-0"
            >
                <RevisionsPopover
                    repo={repo.id}
                    repoName={repo.name}
                    defaultBranch={defaultBranch}
                    currentRev={currentRevision}
                    currentCommitID={currentRevision}
                    togglePopover={togglePopover}
                    getPathFromRevision={getPathFromRevision}
                    showSpeculativeResults={true}
                    onSelect={handleSelect}
                />
            </Popover>
        </Button>
    )
}
