import React from 'react'
import { Link } from '@sourcegraph/wildcard'
import { isExternalURL } from '../utils'

interface LinkOrAnchorProps {
    href: string
    className?: string
    onClick?: (event: React.MouseEvent<HTMLAnchorElement>) => void
}
export const LinkOrAnchor: React.FunctionComponent<LinkOrAnchorProps> = ({ href, children, ...props }) => (
    <Link to={href} {...props} {...(isExternalURL(href) && { target: '_blank', rel: 'noopener noreferrer' })}>
        {children}
    </Link>
)
