import * as React from 'react'
import { Link } from 'react-router-dom'

import { AnchorLink, AnchorLinkProps } from '../AnchorLink'

export const RouterLink: React.FunctionComponent<AnchorLinkProps> = React.forwardRef(
    ({ to, children, ...rest }: AnchorLinkProps, reference) => (
        <AnchorLink to={to} as={Link} {...rest} ref={reference}>
            {children}
        </AnchorLink>
    )
)
