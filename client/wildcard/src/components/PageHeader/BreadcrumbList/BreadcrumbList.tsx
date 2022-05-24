import React from 'react'

import classNames from 'classnames'

import { Heading } from '../../Typography/Heading/Heading'

import styles from './BreadcrumbList.module.scss'

export const BreadcrumbList = React.forwardRef(({ children, as = 'h1', className, ...props }, ref) => (
    <Heading as={as} className={classNames(styles.list, className)} {...props} ref={ref}>
        {children}
    </Heading>
)) as typeof Heading
