import React from 'react'

import classNames from 'classnames'

import type { ForwardReferenceComponent } from '../..'

import type { BUTTON_GROUP_DIRECTION } from './constants'

import styles from './Button.module.scss'

export interface ButtonGroupProps {
    /**
     * Used to change the element that is rendered, default to div
     */
    as?: React.ElementType
    /**
     * Defines the orientaion contained button elements. defaults to horizontal
     */
    direction?: typeof BUTTON_GROUP_DIRECTION[number]
}

export const ButtonGroup = React.forwardRef(
    ({ as: Component = 'div', children, className, direction, ...attributes }, reference) => (
        <Component
            ref={reference}
            role="group"
            className={classNames(styles.btnGroup, direction === 'vertical' && styles.btnGroupVertical, className)}
            {...attributes}
        >
            {children}
        </Component>
    )
) as ForwardReferenceComponent<'div', ButtonGroupProps>
