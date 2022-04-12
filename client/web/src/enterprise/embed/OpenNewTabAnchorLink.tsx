import * as React from 'react'

// eslint-disable-next-line no-restricted-imports
import { Link } from 'react-router-dom'

import { ForwardReferenceComponent, AnchorLink, AnchorLinkProps } from '@sourcegraph/wildcard'

export const OpenNewTabAnchorLink = React.forwardRef(({ children, ...rest }, reference) => (
    <AnchorLink ref={reference} {...rest} target="_blank" rel="noopener noreferrer">
        {children}
    </AnchorLink>
)) as ForwardReferenceComponent<Link<unknown>, AnchorLinkProps>
