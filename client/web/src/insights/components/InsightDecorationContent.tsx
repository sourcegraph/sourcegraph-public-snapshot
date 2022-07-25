import { forwardRef, PropsWithChildren } from 'react'

import styles from './InsightDecorationContent.module.scss'

export const InsightDecorationContent = forwardRef<HTMLSpanElement, PropsWithChildren<{}>>(({ children }, ref) => (
    <span ref={ref} className={styles.insightDecorationContent}>
        {children}
    </span>
))

InsightDecorationContent.displayName = 'InsightDecorationContent'
