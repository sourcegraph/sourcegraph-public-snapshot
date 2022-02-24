import classNames from 'classnames'
import type { MdiReactIconProps } from 'mdi-react'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import React, { HTMLAttributes } from 'react'

import { Icon } from '@sourcegraph/wildcard'

import styles from './RepoRevision.module.scss'

type RepoRevisionProps = HTMLAttributes<HTMLDivElement>

export const RepoRevisionWrapper: React.FunctionComponent<RepoRevisionProps> = ({ children, className, ...rest }) => (
    <div className={classNames(styles.repoRevisionContainer, className)} {...rest}>
        {children}
    </div>
)

export const RepoRevisionChevronDownIcon: React.FunctionComponent<MdiReactIconProps> = ({
    className,
    ref,
    ...rest
}) => <Icon as={ChevronDownIcon} inline={false} className={classNames(styles.breadcrumbIcon, className)} {...rest} />
