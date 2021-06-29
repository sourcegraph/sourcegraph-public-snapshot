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
    onClick?: () => void
}

export interface StepsProps {
    current: number
    children: React.ReactElement<StepProps> | React.ReactElement<StepProps>[]
    numbered?: boolean
    onTabClick?: (index: number) => void
}

export const Step: React.FunctionComponent<StepProps> = ({ title, active, borderColor, disabled, onClick }) => (
    <li
        role="presentation"
        className={classNames(
            disabled && stepsStyles.disabled,
            stepsStyles.listItem,
            active && stepsStyles.active,
            borderColor && stepsStyles[`color${upperFirst(borderColor)}` as keyof typeof stepsStyles]
        )}
        onClick={onClick}
        aria-current={active}
    >
        {title}
    </li>
)

export const Steps: React.FunctionComponent<StepsProps> = ({ current = 1, numbered, onTabClick, children }) => {
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
            // TODO: check this again
            onClick: () => (onTabClick ? onTabClick(index + 1) : undefined),
        })
    })

    return (
        <nav className={stepsStyles.stepsWrapper} aria-label="progress">
            {numbered ? <ol>{element}</ol> : <ul>{element}</ul>}
        </nav>
    )
}
