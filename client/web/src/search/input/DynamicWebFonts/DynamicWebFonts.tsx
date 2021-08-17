import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import styles from './DynamicWebFonts.module.scss'
import { useDynamicWebFonts, DynamicWebFont } from './useDynamicWebFonts'

interface DynamicWebFontsProps {
    fonts: DynamicWebFont[]
}

/**
 * Use native CSS Font Loading Module API to load fonts dynamically.
 * Show loading spinner until fonts are ready.
 */
export const DynamicWebFonts: React.FunctionComponent<DynamicWebFontsProps> = props => {
    const { children, fonts } = props
    const areFontsLoaded = useDynamicWebFonts(fonts)

    // While fonts are not ready, show loading spinner to avoid content jumps.
    if (!areFontsLoaded) {
        return <LoadingSpinner className={styles.spinner} />
    }

    return <>{children}</>
}
