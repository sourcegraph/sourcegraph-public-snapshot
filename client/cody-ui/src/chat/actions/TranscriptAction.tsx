import React, { useState } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'
import classNames from 'classnames'

import styles from './TranscriptAction.module.css'

export interface TranscriptActionStep {
    verb: string
    object: string | JSX.Element

    /**
     * The SVG path of an icon.
     *
     * @example mdiSearchWeb
     */
    icon?: string
}

export const TranscriptAction: React.FunctionComponent<{
    title: string | { verb: string; object: string }
    steps: TranscriptActionStep[]
    className?: string
}> = ({ title, steps, className }) => {
    const [open, setOpen] = useState(false)

    return (
        <div className={classNames(className, styles.container, open && styles.containerOpen)}>
            <button type="button" onClick={() => setOpen(!open)} className={styles.openCloseButton}>
                {typeof title === 'string' ? (
                    title
                ) : (
                    <span>
                        {title.verb} <strong>{title.object}</strong>
                    </span>
                )}
                <Icon
                    aria-hidden={true}
                    svgPath={open ? mdiChevronUp : mdiChevronDown}
                    className={styles.openCloseIcon}
                />
            </button>
            {open && (
                <ol className={styles.steps}>
                    {steps.map((step, index) => (
                        // eslint-disable-next-line react/no-array-index-key
                        <li key={index} className={styles.step}>
                            {step.icon && <Icon svgPath={step.icon} className={styles.stepIcon} />} {step.verb}{' '}
                            <span className={styles.stepObject}>{step.object}</span>
                        </li>
                    ))}
                </ol>
            )}
        </div>
    )
}

const Icon: React.FunctionComponent<{ svgPath: string; className?: string }> = ({ svgPath, className }) => (
    <svg role="img" height={24} width={24} viewBox="0 0 24 24" fill="currentColor" className={className}>
        <path d={svgPath} />
    </svg>
)
