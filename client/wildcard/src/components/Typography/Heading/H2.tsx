import React from 'react'

import type { ForwardReferenceComponent } from '../../../types'

import { Heading, type HeadingProps } from './Heading'

type H2Props = HeadingProps

export const H2 = React.forwardRef(({ children, as = 'h2', ...props }, reference) => (
    <Heading as={as} styleAs="h2" {...props} ref={reference}>
        {children}
    </Heading>
)) as ForwardReferenceComponent<'h2', H2Props>
