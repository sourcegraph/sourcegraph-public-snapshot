import React from 'react'

import type { ForwardReferenceComponent } from '../../../types'

import { Heading, type HeadingProps } from './Heading'

type H6Props = HeadingProps

export const H6 = React.forwardRef(({ children, as = 'h6', ...props }, reference) => (
    <Heading as={as} styleAs="h6" {...props} ref={reference}>
        {children}
    </Heading>
)) as ForwardReferenceComponent<'h6', H6Props>
