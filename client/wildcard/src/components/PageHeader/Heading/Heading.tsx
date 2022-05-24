import React from 'react'

import classNames from 'classnames'

import { Typography } from '../../Typography'

import styles from './Heading.module.scss'

export const Heading = React.forwardRef(({ children, as = 'h1', className, ...props }, reference) => (
    <Typography.Heading as={as} className={classNames(styles.heading, className)} {...props} ref={reference}>
        {children}
    </Typography.Heading>
)) as typeof Typography.Heading
