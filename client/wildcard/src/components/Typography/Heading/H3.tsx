import React from 'react'

import { ForwardReferenceComponent } from '../../../types'

import { Heading, HeadingProps } from './Heading'

type H3Props = HeadingProps

// eslint-disable-next-line id-length
export const H3 = React.forwardRef(({ children, as = 'h3', ...props }, reference) => (
    <Heading as={as} styleAs="h3" {...props} ref={reference}>
        {children}
    </Heading>
)) as ForwardReferenceComponent<'h3', H3Props>
