import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'
import type { MdiReactIconProps } from 'mdi-react'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'

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

export const RepoRevisionChevronDownIcon: React.FunctionComponent<React.PropsWithChildren<MdiReactIconProps>> = ({
    className,
    ...rest
}) => <ChevronDownIcon className={classNames(styles.breadcrumbIcon, className)} {...rest} />
