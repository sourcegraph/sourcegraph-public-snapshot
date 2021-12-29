import isAbsoluteUrl from 'is-absolute-url'
import * as React from 'react'
import { Link } from 'react-router-dom'

import { AnchorLink, LinkProps } from '../AnchorLink'

export const RouterLink: React.FunctionComponent<LinkProps> = ({ to, children, ...rest }) => (
    <AnchorLink to={to} as={typeof to === 'string' && isAbsoluteUrl(to) ? undefined : Link} {...rest}>
        {children}
    </AnchorLink>
)
