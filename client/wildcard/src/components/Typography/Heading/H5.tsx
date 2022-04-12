import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../../types'

import { Heading, HeadingProps } from './Heading'

import styles from './Heading.module.scss'

type H5Props = HeadingProps

// eslint-disable-next-line id-length
export const H5 = React.forwardRef(({ children, as = 'h5', className, ...props }, reference) => (
    <Heading as={as} className={classNames(styles.h5, className)} {...props} ref={reference}>
        {children}
    </Heading>
)) as ForwardReferenceComponent<'h5', H5Props>
