import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../../types'

import { Heading, HeadingProps } from './Heading'

import styles from './Heading.module.scss'

type H1Props = HeadingProps

// eslint-disable-next-line id-length
export const H1 = React.forwardRef(({ children, as = 'h1', className, ...props }, reference) => (
    <Heading as={as} className={classNames(styles.h1, className)} {...props} ref={reference}>
        {children}
    </Heading>
)) as ForwardReferenceComponent<'h1', H1Props>
