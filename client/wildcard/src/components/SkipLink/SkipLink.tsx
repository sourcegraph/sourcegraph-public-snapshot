import React from 'react'

import { Button } from '../Button'
import { AnchorLink } from '../Link'

import styles from './SkipLink.module.scss'

export interface SkipLinkProps {
    href: string
}

/**
 * Skip links
 */
export const SkipLink: React.FunctionComponent<React.PropsWithChildren<SkipLinkProps>> = ({ children, href }) => (
    <Button as={AnchorLink} to={href} className={styles.skipLink} variant="primary">
        {children}
    </Button>
)
