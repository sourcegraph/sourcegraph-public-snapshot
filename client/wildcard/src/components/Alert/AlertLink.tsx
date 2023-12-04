import React from 'react'

import classNames from 'classnames'

import { Link, type LinkProps } from '../Link/Link'

import styles from './AlertLink.module.scss'

export interface AlertLinkProps extends LinkProps {}

export const AlertLink: React.FunctionComponent<React.PropsWithChildren<AlertLinkProps>> = ({
    to,
    children,
    className,
    ...attributes
}) => (
    <Link to={to} className={classNames(styles.alertLink, className)} {...attributes}>
        {children}
    </Link>
)
