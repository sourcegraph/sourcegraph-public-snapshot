import classNames from 'classnames'
import React from 'react'

import { AnchorLink, LinkProps } from '@sourcegraph/wildcard/src/components/Link/AnchorLink'

import styles from './AlertLink.module.scss'

interface AlertLinkProps extends LinkProps {}

export const AlertLink: React.FunctionComponent<AlertLinkProps> = ({ to, children, className, ...attributes }) => (
    <AnchorLink to={to} className={classNames(styles.alertLink, className)} {...attributes}>
        {children}
    </AnchorLink>
)
