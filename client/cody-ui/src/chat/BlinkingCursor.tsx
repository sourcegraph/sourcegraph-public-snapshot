import React from 'react'

import styles from './BlinkingCursor.module.css'

export const BlinkingCursor: React.FunctionComponent = () => (
    <p>
        Working on it
        <span className={styles.blinkDots} />
    </p>
)
