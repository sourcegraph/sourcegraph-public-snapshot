import classNames from 'classnames'
import React, { HTMLAttributes } from 'react'
import { Link, LinkProps } from 'react-router-dom'

import styles from './TreeRowLabel.module.scss'

type TreeRowLabelProps = HTMLAttributes<HTMLSpanElement>

export const TreeRowLabel: React.FunctionComponent<TreeRowLabelProps> = ({ className, children, ...rest }) => (
    <span className={classNames(className, styles.rowLabel)} {...rest}>
        {children}
    </span>
)

type TreeRowLabelLinkProps = LinkProps<any>

export const TreeRowLabelLink: React.FunctionComponent<TreeRowLabelLinkProps> = ({ className, children, ...rest }) => (
    <Link className={classNames(className, styles.rowLabel)} {...rest}>
        {children}
    </Link>
)
