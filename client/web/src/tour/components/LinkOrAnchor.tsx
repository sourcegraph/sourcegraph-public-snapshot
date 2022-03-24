import React from 'react'

import { isExternalLink } from '@sourcegraph/common'
import { Link } from '@sourcegraph/wildcard'

interface LinkOrAnchorProps {
    href: string
    className?: string
    onClick?: (event: React.MouseEvent<HTMLAnchorElement>) => void
}

/**
 * Link component opens external links in a new tab.
 */
export const LinkOrAnchor: React.FunctionComponent<LinkOrAnchorProps> = ({ href, children, ...props }) => (
    <Link to={href} {...props} {...(isExternalLink(href) && { target: '_blank', rel: 'noopener noreferrer' })}>
        {children}
    </Link>
)
