import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../../types'

import { Heading, HeadingProps } from './Heading'

import styles from './Heading.module.scss'

type H3Props = HeadingProps

// eslint-disable-next-line id-length
export const H3 = React.forwardRef(({ children, as = 'h3', className, ...props }, reference) => (
    <Heading as={as} className={classNames(styles.h3, className)} {...props} ref={reference}>
        {children}
    </Heading>
)) as ForwardReferenceComponent<'h3', H3Props>
