import classNames from 'classnames'
import React from 'react'

import { Link, LinkProps } from '@sourcegraph/shared/src/components/Link'

import styles from './AlertLink.module.scss'

interface AlertLinkProps extends LinkProps {}

export const AlertLink: React.FunctionComponent<AlertLinkProps> = ({ to, children, className, ...attributes }) => (
    <Link to={to} className={classNames(styles.alertLink, className)} {...attributes}>
        {children}
    </Link>
)
