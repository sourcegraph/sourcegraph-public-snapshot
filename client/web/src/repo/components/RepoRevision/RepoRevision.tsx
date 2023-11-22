import React, { type HTMLAttributes } from 'react'

import { mdiChevronDown } from '@mdi/js'
import classNames from 'classnames'

import { Icon, type IconProps } from '@sourcegraph/wildcard'

import styles from './RepoRevision.module.scss'

type RepoRevisionProps = HTMLAttributes<HTMLDivElement>

export const RepoRevisionWrapper: React.FunctionComponent<React.PropsWithChildren<RepoRevisionProps>> = ({
    children,
    className,
    ...rest
}) => (
    <div className={classNames(styles.repoRevisionContainer, className)} {...rest}>
        {children}
    </div>
)

export const RepoRevisionChevronDownIcon: React.FunctionComponent<React.PropsWithChildren<IconProps>> = ({
    className,
    ...props
}) => <Icon className={classNames(styles.breadcrumbIcon, className)} svgPath={mdiChevronDown} {...props} />
