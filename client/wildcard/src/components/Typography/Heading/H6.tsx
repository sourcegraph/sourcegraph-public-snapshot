import React from 'react'

import { ForwardReferenceComponent } from '../../../types'

import { Heading, HeadingProps } from './Heading'

type H6Props = HeadingProps

// eslint-disable-next-line id-length
export const H6 = React.forwardRef(({ children, as = 'h6', styleAs = 'h6', ...props }, reference) => (
    <Heading as={as} styleAs={styleAs} {...props} ref={reference}>
        {children}
    </Heading>
)) as ForwardReferenceComponent<'h6', H6Props>
