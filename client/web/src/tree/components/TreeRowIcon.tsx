import React, { AnchorHTMLAttributes, HTMLAttributes } from 'react'

import classNames from 'classnames'

import { Link } from '@sourcegraph/wildcard'

import styles from './TreeRowIcon.module.scss'

type TreeRowIconProps = HTMLAttributes<HTMLSpanElement>

export const TreeRowIcon: React.FunctionComponent<React.PropsWithChildren<TreeRowIconProps>> = ({
    className,
    children,
    ...rest
}) => (
    <span className={classNames(className, styles.rowIcon)} data-testid="tree-row-icon" {...rest}>
        {children}
    </span>
)

type TreeRowIconLinkProps = AnchorHTMLAttributes<HTMLAnchorElement>

export const TreeRowIconLink: React.FunctionComponent<React.PropsWithChildren<TreeRowIconLinkProps>> = ({
    className,
    children,
    href,
    ...rest
}) => (
    <Link
        to={href !== undefined ? href : ''}
        className={classNames(className, styles.rowIcon)}
        data-testid="tree-row-icon"
        {...rest}
    >
        {children}
    </Link>
)
