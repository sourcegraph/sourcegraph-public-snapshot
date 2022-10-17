import React from 'react'

import { ForwardReferenceComponent } from '../../../types'

import { Heading, HeadingProps } from './Heading'

type H1Props = HeadingProps

// eslint-disable-next-line id-length
export const H1 = React.forwardRef(({ children, as = 'h1', ...props }, reference) => (
    <Heading as={as} styleAs="h1" {...props} ref={reference}>
        {children}
    </Heading>
)) as ForwardReferenceComponent<'h1', H1Props>
