import React, { useEffect, useState } from 'react'
import { getSemanticColorVariables } from '../utils'
import styles from './ColorVariants.module.scss'

export const ColorVariants: React.FunctionComponent = () => {
    const [colors, setColors] = useState<string[][]>([])
    useEffect(() => {
        setTimeout(() => {
            setColors(getSemanticColorVariables())
        })
    }, [])

    return (
        <div className={styles.grid}>
            {colors.map(colors => {
                return (
                    <>
                        {colors.map(color => (
                            <div className="m-2 text-center" key={color}>
                                <div
                                    className="rounded"
                                    style={{ width: '6rem', height: '6rem', backgroundColor: `var(${color})` }}
                                />
                                <code>{color}</code>
                            </div>
                        ))}
                    </>
                )
            })}
        </div>
    )
}
