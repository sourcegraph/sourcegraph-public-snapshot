import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import * as React from 'react'
import { Popover } from 'reactstrap'

import { escapeRevspecForURL } from '@sourcegraph/shared/src/util/url'
import { Button, PageHeader } from '@sourcegraph/wildcard'

import { RevisionsPopover } from '../revisions-popover/RevisionsPopover'

import { RepositoryCompareAreaPageProps } from './RepositoryCompareArea'

interface RepositoryCompareHeaderProps extends RepositoryCompareAreaPageProps {
    className: string

    /** Called when the user updates the comparison spec and submits the form. */
    onUpdateComparisonSpec: (newBaseSpec: string, newHeadSpec: string) => void
}

interface RevisionComparison {
    base: RepositoryCompareHeaderProps['base']
    head: RepositoryCompareHeaderProps['head']
}

export const RepositoryComparePopover: React.FunctionComponent<{
    /**
     * Uniquely identify this specific popover. Used to link the trigger button with the popover to display
     */
    id: string
    /**
     * Initial revision comparison to load. This can be changed by interactiing with the popover.
     */
    comparison: RevisionComparison
    /**
     * The specific comparison type that the popover is concerned with changing.
     */
    type: keyof RevisionComparison
    repo: RepositoryCompareHeaderProps['repo']
}> = ({ id, comparison, repo, type }) => {
    const [popoverOpen, setPopoverOpen] = React.useState(false)
    const togglePopover = React.useCallback(() => setPopoverOpen(previous => !previous), [])

    /**
     * Override the default node URL behavior to support navigating to a repository sub-page.
     */
    const getURLFromRevision = React.useCallback(
        (_href: string, revision: string) => {
            const escapedRevision = escapeRevspecForURL(revision)
            const comparePath =
                type === 'base'
                    ? `${escapedRevision}...${escapeRevspecForURL(comparison.head.revision || '')}`
                    : `${escapeRevspecForURL(comparison.base.revision || '')}...${escapedRevision}`

            return `/${repo.name}/-/compare/${comparePath}`
        },
        [comparison, repo.name, type]
    )

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
            {comparison[type].revision || repo.defaultBranch?.abbrevName || 'HEAD'}
            <ChevronDownIcon className="icon-inline repo-revision-container__breadcrumb-icon" />
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
                    // TODO: abbrevName? or displayName?
                    defaultBranch={repo.defaultBranch?.abbrevName || ''}
                    // TODO support currentRev? Or default to branch
                    currentRev={undefined}
                    togglePopover={togglePopover}
                    getURLFromRevision={getURLFromRevision}
                    allowSpeculativeSearch={true}
                />
            </Popover>
        </Button>
    )
}

export const RepositoryCompareHeader: React.FunctionComponent<RepositoryCompareHeaderProps> = ({
    base,
    head,
    className,
    repo,
}) => (
    <div className={`repository-compare-header ${className}`}>
        <PageHeader
            path={[{ text: 'Compare changes' }]}
            description="Compare changes across revisions."
            className="mb-3"
        />
        <div className={`${className}-inner d-flex align-items-center`}>
            <RepositoryComparePopover id="base-popover" type="base" comparison={{ base, head }} repo={repo} />
            <DotsHorizontalIcon className="icon-inline mx-2" />
            <RepositoryComparePopover id="head-popover" type="head" comparison={{ base, head }} repo={repo} />
        </div>
    </div>
)
