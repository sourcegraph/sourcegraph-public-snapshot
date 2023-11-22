import React from 'react'

import { PerformanceGauge, type Props } from '../PerformanceGauge'

import styles from './StyledPerformanceGauge.module.scss'

export const StyledPerformanceGauge: React.FunctionComponent<
    React.PropsWithChildren<Exclude<Props, 'countClassName' | 'labelClassName'>>
> = props => <PerformanceGauge countClassName={styles.count} labelClassName={styles.label} {...props} />
