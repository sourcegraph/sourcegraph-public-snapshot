import React from 'react'

import classNames from 'classnames'

import { Heading as TypographyHeading } from '../../Typography'

import styles from './Heading.module.scss'

export const Heading = React.forwardRef(({ children, as = 'h1', className, ...props }, reference) => (
    <TypographyHeading as={as} className={classNames(styles.heading, className)} {...props} ref={reference}>
        {children}
    </TypographyHeading>
)) as typeof TypographyHeading
