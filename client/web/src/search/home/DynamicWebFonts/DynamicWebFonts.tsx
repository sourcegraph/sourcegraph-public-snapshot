import React from 'react'

import { LoadingSpinner } from '@sourcegraph/wildcard'

import styles from './DynamicWebFonts.module.scss'
import { useDynamicWebFonts, DynamicWebFont } from './useDynamicWebFonts'

interface DynamicWebFontsProps {
    fonts: DynamicWebFont[]
}

/**
 * Use native CSS Font Loading Module API to load fonts dynamically.
 * Show loading spinner until fonts are ready.
 * In case of a network error proceed to UI rendering.
 */
export const DynamicWebFonts: React.FunctionComponent<DynamicWebFontsProps> = props => {
    const { children, fonts } = props
    const areFontsLoading = useDynamicWebFonts(fonts)

    // While fonts are not ready, show loading spinner to avoid content jumps.
    if (areFontsLoading) {
        return <LoadingSpinner className={styles.spinner} />
    }

    return <>{children}</>
}
