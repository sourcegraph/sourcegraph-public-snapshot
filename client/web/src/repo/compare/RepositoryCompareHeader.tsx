import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import * as React from 'react'
import { useHistory, useLocation } from 'react-router'
import { Popover } from 'reactstrap'

import { Button } from '@sourcegraph/wildcard'

import { BRANCHES_TAB, RevisionsPopover, TAGS_TAB } from '../RevisionsPopover'

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
    id: string
    comparison: RevisionComparison
    type: keyof RevisionComparison
    repo: RepositoryCompareHeaderProps['repo']
}> = ({ id, comparison, repo, type }) => {
    const history = useHistory()
    const location = useLocation()
    const [popoverOpen, setPopoverOpen] = React.useState(false)
    const togglePopover = React.useCallback(() => setPopoverOpen(previous => !previous), [])

    const getURLFromRevision = React.useCallback(
        (href: string, revision: string) => {
            // TODO: URL handling
            const comparePath =
                type === 'base'
                    ? `${revision}...${comparison.head.revision || ''}`
                    : `${comparison.base.revision || ''}...${revision}`
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
                    defaultBranch={repo.defaultBranch?.abbrevName || ''}
                    currentRev={undefined}
                    history={history}
                    location={location}
                    togglePopover={togglePopover}
                    getURLFromRevision={getURLFromRevision}
                    tabs={[BRANCHES_TAB, TAGS_TAB]}
                />
            </Popover>
        </Button>
    )
}

/**
 * Header for the repository compare area.
 */
export const RepositoryCompareHeader: React.FunctionComponent<RepositoryCompareHeaderProps> = ({
    base,
    head,
    className,
    repo,
}) => (
    <div className={`repository-compare-header ${className}`}>
        <div className={`${className}-inner d-flex align-items-center`}>
            <RepositoryComparePopover id="base-popover" type="base" comparison={{ base, head }} repo={repo} />
            <DotsHorizontalIcon className="icon-inline mx-2" />
            <RepositoryComparePopover id="head-popover" type="head" comparison={{ base, head }} repo={repo} />
        </div>
    </div>
)
