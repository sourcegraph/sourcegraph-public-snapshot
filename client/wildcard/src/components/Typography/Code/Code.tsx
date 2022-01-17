import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../../types'
import typographyStyles from '../Typography.module.scss'
import { getFontWeightStyle } from '../utils'

import styles from './Code.module.scss'

interface CodeProps extends React.HTMLAttributes<HTMLElement> {
    size?: 'small' | 'base'
    weight?: 'regular' | 'medium' | 'bold'
}

export const Code = React.forwardRef(({ children, as: Component = 'code', size, weight, className }, reference) => (
    <Component
        className={classNames(
            styles.code,
            size === 'small' && typographyStyles.small,
            weight && getFontWeightStyle({ weight }),
            className
        )}
        ref={reference}
    >
        {children}
    </Component>
)) as ForwardReferenceComponent<'code', CodeProps>
