/* eslint-disable react/forbid-dom-props */
import React, { useState, useEffect } from 'react'

import { getSemanticColorVariables, SemanticColor } from '../utils'

import styles from './ColorVariants.module.scss'

export const ColorVariants: React.FunctionComponent = () => {
    const [colors, setColors] = useState<SemanticColor[]>()

    // Styles are loaded after initial render
    useEffect(() => {
        setColors(getSemanticColorVariables())
    }, [])

    return (
        <div className={styles.grid}>
            {colors?.map(variant => (
                <div className="m-2 text-center" key={variant}>
                    <div
                        className="rounded"
                        style={{ width: '6rem', height: '6rem', backgroundColor: `var(${variant})` }}
                    />
                    <code>{variant}</code>
                </div>
            ))}
        </div>
    )
}
