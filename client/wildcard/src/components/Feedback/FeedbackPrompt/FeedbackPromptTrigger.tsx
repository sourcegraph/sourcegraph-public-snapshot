import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent, PopoverTrigger } from '../../..'
import { PopoverTriggerProps } from '../../Popover'

import styles from './FeedbackPromptTrigger.module.scss'

export const FeedbackPromptTrigger = React.forwardRef(({ children, className, ...otherAttributes }, reference) => (
    <PopoverTrigger ref={reference} className={classNames(className, styles.toggle)} {...otherAttributes}>
        {children}
    </PopoverTrigger>
)) as ForwardReferenceComponent<'button', PopoverTriggerProps>
