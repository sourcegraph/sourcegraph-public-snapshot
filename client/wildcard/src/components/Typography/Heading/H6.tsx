import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../../types'

import { Heading, HeadingProps } from './Heading'

import styles from './Heading.module.scss'

type H6Props = HeadingProps

// eslint-disable-next-line id-length
export const H6 = React.forwardRef(({ children, as = 'h6', className, ...props }, reference) => (
    <Heading as={as} className={classNames(styles.h6, className)} {...props} ref={reference}>
        {children}
    </Heading>
)) as ForwardReferenceComponent<'h6', H6Props>
