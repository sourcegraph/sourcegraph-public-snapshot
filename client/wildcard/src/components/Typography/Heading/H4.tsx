import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../../types'

import { Heading, HeadingProps } from './Heading'

import styles from './Heading.module.scss'

type H4Props = HeadingProps

// eslint-disable-next-line id-length
export const H4 = React.forwardRef(({ children, as = 'h4', className, ...props }, reference) => (
    <Heading as={as} className={classNames(styles.h4, className)} {...props} ref={reference}>
        {children}
    </Heading>
)) as ForwardReferenceComponent<'h4', H4Props>
