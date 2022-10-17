import { forwardRef } from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '@sourcegraph/wildcard'

import styles from './TruncatedText.module.scss'

export const TruncatedText = forwardRef((props, reference) => {
    const { as: Component = 'span', className, ...otherProps } = props

    return <Component className={classNames(className, styles.truncatedText)} {...otherProps} />
}) as ForwardReferenceComponent<'span'>
