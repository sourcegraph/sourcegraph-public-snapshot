import React from 'react'

export interface StepProps {
    title: React.ReactNode
    description?: React.ReactNode
    active?: boolean
}

export interface StepsProps {
    current: number
    initial: number
    numbered?: boolean
    onChange: (current: number) => void
    children?: React.ReactNode
}

export const Step: React.FunctionComponent<StepProps> = ({ title, active }) => <li aria-current={active}>{title}</li>

export const Steps: React.FunctionComponent<StepsProps> = ({ children, numbered, current = 1 }) => {
    if (!children) {
        throw new Error('Steps must be include at least one child')
    }

    console.log(current, React.Children.count(children))

    if (current < 1 || current > React.Children.count(children)) {
        console.warn('this is out of the limits')
    }

    const element = React.Children.map(children, (child, index) => {
        const active = current === index + 1
        return React.cloneElement(child as React.ReactElement, { active })
    })

    return <nav aria-label="progress">{numbered ? <ol>{element}</ol> : <ul>{element}</ul>}</nav>
}
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
