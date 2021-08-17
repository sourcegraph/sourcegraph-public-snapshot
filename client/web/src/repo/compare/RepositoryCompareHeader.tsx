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
            path={[{ text: 'Compare changes' }]}
            description="Compare changes across revisions."
            className="mb-3"
        />
        <div className="d-flex align-items-center">
            <RepositoryComparePopover id="base-popover" type="base" comparison={{ base, head }} repo={repo} />
            <DotsHorizontalIcon className="icon-inline mx-2" />
            <RepositoryComparePopover id="head-popover" type="head" comparison={{ base, head }} repo={repo} />
        </div>
    </div>
)
