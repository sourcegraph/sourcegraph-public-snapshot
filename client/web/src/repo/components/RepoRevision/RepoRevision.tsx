import React, { HTMLAttributes } from 'react'

import { mdiChevronDown } from '@mdi/js'
import classNames from 'classnames'

import { Icon, IconProps } from '@sourcegraph/wildcard'

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

export const RepoRevisionChevronDownIcon: React.FunctionComponent<
    React.PropsWithChildren<React.PropsWithoutRef<IconProps>>
> = ({ className, ...rest }) => (
    <Icon
        className={classNames(styles.breadcrumbIcon, className)}
        {...rest}
        svgPath={mdiChevronDown}
        inline={false}
        aria-hidden={true}
    />
)
