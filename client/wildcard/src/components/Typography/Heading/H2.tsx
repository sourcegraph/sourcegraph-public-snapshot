import React from 'react'

import { ForwardReferenceComponent } from '../../../types'

import { Heading, HeadingProps } from './Heading'

type H2Props = HeadingProps

// eslint-disable-next-line id-length
export const H2 = React.forwardRef(({ children, as = 'h2', styleAs = 'h2', ...props }, reference) => (
    <Heading as={as} styleAs={styleAs} {...props} ref={reference}>
        {children}
    </Heading>
)) as ForwardReferenceComponent<'h2', H2Props>
