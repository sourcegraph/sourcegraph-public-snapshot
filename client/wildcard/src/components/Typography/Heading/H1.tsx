import React from 'react'

import type { ForwardReferenceComponent } from '../../../types'

import { Heading, type HeadingProps } from './Heading'

type H1Props = HeadingProps

export const H1 = React.forwardRef(({ children, as = 'h1', ...props }, reference) => (
    <Heading as={as} styleAs="h1" {...props} ref={reference}>
        {children}
    </Heading>
)) as ForwardReferenceComponent<'h1', H1Props>
