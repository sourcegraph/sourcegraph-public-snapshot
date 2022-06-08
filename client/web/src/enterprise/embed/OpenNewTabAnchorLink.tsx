import * as React from 'react'

import { ForwardReferenceComponent, AnchorLink, LinkProps } from '@sourcegraph/wildcard'

export const OpenNewTabAnchorLink = React.forwardRef(({ children, ...rest }, reference) => (
    <AnchorLink ref={reference} {...rest} target="_blank" rel="noopener noreferrer">
        {children}
    </AnchorLink>
)) as ForwardReferenceComponent<'a', LinkProps>
