import classNames from 'classnames'
import { upperFirst } from 'lodash'
import React from 'react'

import stepsStyles from './Steps.module.scss'

export interface StepProps {
    title: React.ReactNode
    description?: React.ReactNode
    active?: boolean
    borderColor?: 'orange' | 'blue' | 'purple'
}

export interface StepsProps {
    current: number
    numbered?: boolean
    // onChange: (current: number) => void
    children?: React.ReactElement<StepProps> | React.ReactElement<StepProps>[]
}

export const Step: React.FunctionComponent<StepProps> = ({ title, active, borderColor }) => (
    <li
        className={classNames(
            stepsStyles.listItem,
            active && stepsStyles.active,
            borderColor && stepsStyles[`color${upperFirst(borderColor)}`]
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

    const element = React.Children.map(children, (child, index) =>
        React.cloneElement(child as React.ReactElement, { active: current === index + 1 })
    )

    return (
        <nav className={stepsStyles.stepsWrapper} aria-label="progress">
            {numbered ? <ol>{element}</ol> : <ul>{element}</ul>}
        </nav>
    )
}
// A11y inspiration
{
    /* <nav aria-label="progress">
  <ul class="progress-tracker progress-tracker--text progress-tracker--center">
    <li class="progress-step is-complete">
      ...
    </li>
    <li class="progress-step is-complete">
      ...
    </li>
    <li class="progress-step is-active" aria-current="true">
      ...
    </li>
    ...
  </ul>
</nav> */
}
