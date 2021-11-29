import classNames from 'classnames'
import React, { AnchorHTMLAttributes, HTMLAttributes } from 'react'

import styles from './TreeRowIcon.module.scss'

type TreeRowIconProps = HTMLAttributes<HTMLSpanElement>

export const TreeRowIcon: React.FunctionComponent<TreeRowIconProps> = ({ className, children, ...rest }) => (
    <span className={classNames(className, styles.rowIcon)} data-testid="tree-row-icon" {...rest}>
        {children}
    </span>
)

type TreeRowIconLinkProps = AnchorHTMLAttributes<HTMLAnchorElement>

export const TreeRowIconLink: React.FunctionComponent<TreeRowIconLinkProps> = ({ className, children, ...rest }) => (
    <a className={classNames(className, styles.rowIcon)} data-testid="tree-row-icon" {...rest}>
        {children}
    </a>
)
