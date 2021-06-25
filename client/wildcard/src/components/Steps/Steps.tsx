import classNames from 'classnames'
import { upperFirst } from 'lodash'
import React from 'react'

import stepsStyles from './Steps.module.scss'

type Color = 'orange' | 'blue' | 'purple'

export interface StepProps {
    title: React.ReactNode
    description?: React.ReactNode
    active?: boolean
    disabled?: boolean
    borderColor?: Color
}

export interface StepsProps {
    current: number
    numbered?: boolean
    children: React.ReactElement<StepProps> | React.ReactElement<StepProps>[]
}

export const Step: React.FunctionComponent<StepProps> = ({ title, active, borderColor, disabled }) => (
    <li
        className={classNames(
            disabled && stepsStyles.disabled,
            stepsStyles.listItem,
            active && stepsStyles.active,
            borderColor && stepsStyles[`color${upperFirst(borderColor)}` as keyof typeof stepsStyles]
        )}
        aria-current={active}
    >
        {title}
    </li>
)

export const Steps: React.FunctionComponent<StepsProps> = ({ children, numbered, current = 1 }) => {
    if (!children) {
        throw new Error('Steps must include at least one child')
    }

    if (current < 1 || current > React.Children.count(children)) {
        console.warn('current step is out of limits')
    }

    const element = React.Children.map(children, (child: React.ReactElement<StepProps>, index) => {
        if (child.type !== Step) {
            throw new Error(`${child.type.toString()} element is not <Step> component`)
        }

        return React.cloneElement(child, {
            disabled: current < index + 1,
            active: current === index + 1,
        })
    })

    return (
        <nav className={stepsStyles.stepsWrapper} aria-label="progress">
            {numbered ? <ol>{element}</ol> : <ul>{element}</ul>}
        </nav>
    )
}
