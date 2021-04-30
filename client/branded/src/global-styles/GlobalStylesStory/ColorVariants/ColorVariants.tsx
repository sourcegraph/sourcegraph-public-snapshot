import React from 'react'
import openColor from 'open-color'

import { getSemanticColorVariables } from '../utils'
import styles from './ColorVariants.module.scss'
import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'
import { SEMANTIC_COLORS } from '../constants'

export const ColorVariants: React.FunctionComponent = () => {
    const [isRedesignEnabled] = useRedesignToggle()

    if (isRedesignEnabled) {
        return (
            <div className={styles.grid}>
                {getSemanticColorVariables().map(variant => (
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

    return (
        <>
            <div className="d-flex flex-wrap">
                {SEMANTIC_COLORS.map(semantic => (
                    <div className="m-2 text-center" key={semantic}>
                        <div className={`bg-${semantic} rounded`} style={{ width: '5rem', height: '5rem' }} />
                        {semantic}
                    </div>
                ))}
            </div>

            <h2>Color Palette</h2>
            <p>
                Our color palette is the <a href="https://yeun.github.io/open-color/">Open Color</a> palette. All colors
                are available as SCSS and CSS variables. It's generally not advised to use these directly, but they may
                be used in rare cases, like charts. In other cases, rely on CSS components, utilities for borders and
                background, and dynamic CSS variables.
            </p>
            {Object.entries(openColor).map(
                ([name, colors]) =>
                    Array.isArray(colors) && (
                        <div key={name}>
                            <h5>{name}</h5>
                            <div className="d-flex flex-wrap">
                                {colors.map((color, number) => (
                                    <div key={color} className="m-2 text-right">
                                        <div
                                            className="rounded"
                                            style={{ background: color, width: '3rem', height: '3rem' }}
                                        />
                                        {number}
                                    </div>
                                ))}
                            </div>
                        </div>
                    )
            )}
        </>
    )
}
