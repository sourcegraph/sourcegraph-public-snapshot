import { forwardRef } from 'react'

import classNames from 'classnames'

import type { ForwardReferenceComponent } from '@sourcegraph/wildcard'

import styles from './CodeInsightsQueryBlock.module.scss'

interface CodeInsightsQueryBlockProps {}

export const CodeInsightsQueryBlock = forwardRef((props, reference) => {
    const { as: Component = 'span', className, ...otherProps } = props

    return <Component ref={reference} {...otherProps} className={classNames(className, styles.query)} />
}) as ForwardReferenceComponent<'span', CodeInsightsQueryBlockProps>
