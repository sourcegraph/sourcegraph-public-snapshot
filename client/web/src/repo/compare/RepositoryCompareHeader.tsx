import React from 'react'

import classNames from 'classnames'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'

import { PageHeader, Link, Icon } from '@sourcegraph/wildcard'

import { RepositoryCompareAreaPageProps } from './RepositoryCompareArea'
import { RepositoryComparePopover } from './RepositoryComparePopover'

import styles from './RepositoryCompareHeader.module.scss'

interface RepositoryCompareHeaderProps extends RepositoryCompareAreaPageProps {
    className: string
}

export const RepositoryCompareHeader: React.FunctionComponent<
    React.PropsWithChildren<RepositoryCompareHeaderProps>
> = ({ base, head, className, repo }) => (
    <div className={classNames(styles.repositoryCompareHeader, className)}>
        <PageHeader
            path={[{ text: 'Compare changes across revisions' }]}
            description={
                <span className="mb-3">
                    Select a revision or provide a{' '}
                    <Link
                        to="https://git-scm.com/docs/git-rev-parse.html#_specifying_revisions"
                        rel="noopener noreferrer"
                        target="_blank"
                    >
                        Git revspec
                    </Link>{' '}
                    for more fine-grained comparisons
                </span>
            }
        />
        <div className="d-flex align-items-center">
            <RepositoryComparePopover id="base-popover" type="base" comparison={{ base, head }} repo={repo} />
            <Icon role="img" className="mx-2" as={DotsHorizontalIcon} aria-hidden={true} />
            <RepositoryComparePopover id="head-popover" type="head" comparison={{ base, head }} repo={repo} />
        </div>
    </div>
)
