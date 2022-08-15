import React from 'react'

import { ForwardReferenceComponent } from '../../../types'

import { Heading, HeadingProps } from './Heading'

type H4Props = HeadingProps

// eslint-disable-next-line id-length
export const H4 = React.forwardRef(({ children, as = 'h4', ...props }, reference) => (
    <Heading as={as} styleAs="h4" {...props} ref={reference}>
        {children}
    </Heading>
)) as ForwardReferenceComponent<'h4', H4Props>
