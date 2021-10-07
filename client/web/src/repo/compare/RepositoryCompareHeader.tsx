import classNames from 'classnames'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import React from 'react'

import { PageHeader } from '@sourcegraph/wildcard'

import { RepositoryCompareAreaPageProps } from './RepositoryCompareArea'
import styles from './RepositoryCompareHeader.module.scss'
import { RepositoryComparePopover } from './RepositoryComparePopover'

interface RepositoryCompareHeaderProps extends RepositoryCompareAreaPageProps {
    className: string
}

export const RepositoryCompareHeader: React.FunctionComponent<RepositoryCompareHeaderProps> = ({
    base,
    head,
    className,
    repo,
}) => (
    <div className={classNames(styles.repositoryCompareHeader, className)}>
        <PageHeader
            path={[{ text: 'Compare changes across revisions' }]}
            description={
                <p>
                    Select a revision or provide a{' '}
                    <a
                        href="https://git-scm.com/docs/git-rev-parse.html#_specifying_revisions"
                        rel="noopener noreferrer"
                        target="_blank"
                    >
                        Git revspec
                    </a>{' '}
                    for more fine-grained comparisons
                </p>
            }
        />
        <div className="d-flex align-items-center">
            <RepositoryComparePopover id="base-popover" type="base" comparison={{ base, head }} repo={repo} />
            <DotsHorizontalIcon className="icon-inline mx-2" />
            <RepositoryComparePopover id="head-popover" type="head" comparison={{ base, head }} repo={repo} />
        </div>
    </div>
)
