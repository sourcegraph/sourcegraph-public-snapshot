import { forwardRef } from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '@sourcegraph/wildcard'

import styles from './AggregationDataContainer.module.scss'

interface DataLayoutContainerProps {
    size?: 'sm' | 'md'
}

export const DataLayoutContainer = forwardRef(function datayLayoutContainerRef(props, ref) {
    const { as: Component = 'div', size = 'md', className, ...attributes } = props

    return (
        <Component
            {...attributes}
            ref={ref}
            className={classNames(className, styles.errorContainer, {
                [styles.errorContainerSmall]: size === 'sm',
            })}
        />
    )
}) as ForwardReferenceComponent<'div', DataLayoutContainerProps>
