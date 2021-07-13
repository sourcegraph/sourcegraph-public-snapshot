import classNames from 'classnames'
import { upperFirst } from 'lodash'
import React, { useRef, useState, useMemo } from 'react'

import stepsStyles from './Steps.module.scss'

type Color = 'orange' | 'blue' | 'purple'

export interface StepProps {
    borderColor: Color
    children: React.ReactNode
}

interface StepContext {
    current: number
    setCurrent: (update: number | ((previousState: number) => number)) => void
    initialStep: number
}

interface StepsData {
    stepIndex: number
    visited: boolean
}
interface StepListContext {
    current: number
    steps: StepsData[] | []
}

interface StepsContext {
    actions: StepContext
    steps: StepListContext
}

interface StepList {
    numeric: boolean
    children: React.ReactElement<StepProps> | React.ReactElement<StepProps, string | React.JSXElementConstructor<any>>[]
}

export interface StepsProps {
    numeric?: boolean
    children: React.ReactElement<StepProps> | React.ReactElement<StepProps>[]
    initialStep: number
}

const StepContext = React.createContext<StepContext | null>(null)
StepContext.displayName = 'StepContext'

const StepListContext = React.createContext<StepListContext | null>(null)
StepContext.displayName = 'StepListContext'

const useStepListContext = (): StepListContext => {
    const context = React.useContext(StepListContext)
    if (!context) {
        throw new Error('StepList compound components cannot be rendered outside the TODO component')
    }
    return context
}

const useStepsContext = (): StepContext => {
    const context = React.useContext(StepContext)
    if (!context) {
        throw new Error('Steps compound components cannot be rendered outside the <Steps> component')
    }
    return context
}

export const useSteps = (): StepsContext | null => {
    const stepsContext = React.useContext(StepContext)
    const stepListContext = React.useContext(StepListContext)
    console.log(stepsContext, stepListContext)
    if (!stepsContext || !stepListContext) {
        return null
    }

    return { actions: stepsContext, data: stepListContext }
}

export const Step: React.FunctionComponent<StepProps> = ({ children, borderColor }) => {
    const { setCurrent, current } = useStepsContext()
    const context = useStepListContext()

    console.log('context ma frend', context)

    const disabled = current < context[0].stepIndex + 1
    const active = current === context[0].stepIndex + 1
    return (
        <li
            role="presentation"
            className={classNames(
                stepsStyles.cursorPointer,
                disabled && stepsStyles.disabled,
                stepsStyles.listItem,
                active && stepsStyles.active,
                borderColor && stepsStyles[`color${upperFirst(borderColor)}` as keyof typeof stepsStyles]
            )}
            aria-current={active}
            onClick={() => setCurrent(context[0].stepIndex + 1)}
        >
            {children}
        </li>
    )
}

export const Steps: React.FunctionComponent<StepsProps> = ({ initialStep = 1, children }) => {
    const [current, setCurrent] = useState(initialStep)
    // const { current } = useSteps(initialStep, React.Children.toArray(children).length)

    if (!children) {
        throw new Error('Steps must include at least one child')
    }

    if (initialStep < 1 || initialStep > React.Children.count(children)) {
        console.warn('current step is out of limits')
    }

    const value = {
        current,
        setCurrent,
        initialStep,
    }

    return <StepContext.Provider value={value}>{useMemo(() => children, [children])}</StepContext.Provider>
}

export const StepList: React.FunctionComponent<StepList> = ({ children, numeric }) => {
    const { current } = useStepsContext()
    const data = useRef<StepListContext>({ current: 1, [{stepIndex: 1, visited: true}]})
    // let data: { stepIndex: number; visited: boolean }[] | [] = []

    const element = React.Children.map(children, (child: React.ReactElement<StepProps>, index) => {
        if (child.type !== Step) {
            throw new Error(`${child.type.toString()} element is not <Step> component`)
        }

        const visited = current === index

        // setStepData(previous => [...previous, { stepIndex: index, visited }])
        console.log(data.current)
        data.current = { current, steps: [...data.current, { stepIndex: index, visited }] }

        return <StepListContext.Provider value={data.current}>{child}</StepListContext.Provider>
    })

    return (
        <nav className={stepsStyles.stepsWrapper} aria-label="progress">
            {numeric ? <ol>{element}</ol> : <ul>{element}</ul>}
        </nav>
    )
}

export const StepPanels: React.FunctionComponent = ({ children }) => {
    const { current } = useStepsContext()
    if (!children) {
        throw new Error('bum!')
    }

    return <div className="mt-4 pb-3">{React.Children.toArray(children)[current - 1]}</div>
}

export const StepPanel: React.FunctionComponent = ({ children }) => <>{children}</>

export const StepActions: React.FunctionComponent = ({ children }) => <>{children}</>
