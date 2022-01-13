import classNames from 'classnames'
import React, { HTMLAttributes } from 'react'

import { Link } from '@sourcegraph/wildcard'
import type { LinkProps } from '@sourcegraph/wildcard/src/components/Link'

import styles from './TreeRowLabel.module.scss'

type TreeRowLabelProps = HTMLAttributes<HTMLSpanElement>

export const TreeRowLabel: React.FunctionComponent<TreeRowLabelProps> = ({ className, children, ...rest }) => (
    <span className={classNames(className, styles.rowLabel)} data-testid="tree-row-label" {...rest}>
        {children}
    </span>
)

type TreeRowLabelLinkProps = LinkProps

export const TreeRowLabelLink: React.FunctionComponent<TreeRowLabelLinkProps> = ({ className, children, ...rest }) => (
    <Link className={classNames(className, styles.rowLabel)} data-testid="tree-row-label" {...rest}>
        {children}
    </Link>
)
