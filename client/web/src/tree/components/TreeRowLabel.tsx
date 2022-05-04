import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'

import { Link, LinkProps } from '@sourcegraph/wildcard'

import styles from './TreeRowLabel.module.scss'

type TreeRowLabelProps = HTMLAttributes<HTMLSpanElement>

export const TreeRowLabel: React.FunctionComponent<React.PropsWithChildren<TreeRowLabelProps>> = ({
    className,
    children,
    ...rest
}) => (
    <span className={classNames(className, styles.rowLabel)} data-testid="tree-row-label" {...rest}>
        {children}
    </span>
)

type TreeRowLabelLinkProps = LinkProps

export const TreeRowLabelLink: React.FunctionComponent<React.PropsWithChildren<TreeRowLabelLinkProps>> = ({
    className,
    children,
    ...rest
}) => (
    <Link className={classNames(className, styles.rowLabel)} data-testid="tree-row-label" {...rest}>
        {children}
    </Link>
)
