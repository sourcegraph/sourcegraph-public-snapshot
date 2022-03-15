import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../../types'

import { Heading, HeadingProps } from './Heading'

import styles from './Heading.module.scss'

type H2Props = HeadingProps

// eslint-disable-next-line id-length
export const H2 = React.forwardRef(({ children, as = 'h2', className, ...props }, reference) => (
    <Heading as={as} className={classNames(styles.h2, className)} {...props} ref={reference}>
        {children}
    </Heading>
)) as ForwardReferenceComponent<'h2', H2Props>
