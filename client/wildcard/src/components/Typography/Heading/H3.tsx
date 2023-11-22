import React from 'react'

import type { ForwardReferenceComponent } from '../../../types'

import { Heading, type HeadingProps } from './Heading'

type H3Props = HeadingProps

export const H3 = React.forwardRef(function H3({ children, as = 'h3', ...props }, reference) {
    return (
        <Heading as={as} styleAs="h3" {...props} ref={reference}>
            {children}
        </Heading>
    )
}) as ForwardReferenceComponent<'h3', H3Props>
