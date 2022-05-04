/* eslint-disable react/forbid-dom-props */
import React from 'react'

import { getSemanticColorVariables } from '../utils'

import styles from './ColorVariants.module.scss'

export const ColorVariants: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
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
