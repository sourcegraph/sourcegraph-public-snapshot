import React, { useEffect, useMemo, useReducer } from 'react'

import classNames from 'classnames'
import { upperFirst } from 'lodash'

import { StepsContext, useStepsContext, StepListContext, useStepListContext, Steps as StepsInterface } from './context'
import { initialState, reducer } from './reducer'

import stepsStyles from './Steps.module.scss'

type Color = 'orange' | 'blue' | 'purple' | 'green'

export interface StepProps {
    borderColor: Color
    children: React.ReactNode
}

interface StepListProps {
    numeric: boolean
    className?: string
    children: React.ReactElement<StepProps> | React.ReactElement<StepProps, string | React.JSXElementConstructor<any>>[]
}

export interface StepsProps {
    children: React.ReactElement<StepProps> | React.ReactElement<StepProps>[]
    initialStep: number
    totalSteps: number
}

export const Steps: React.FunctionComponent<React.PropsWithChildren<StepsProps>> = ({
    initialStep = 1,
    totalSteps,
    children,
}) => {
    const [state, dispatch] = useReducer(reducer, initialState(initialStep, totalSteps))

    if (!children) {
        throw new Error('Steps must include at least one child')
    }

    if (initialStep < 1 || initialStep > totalSteps) {
        throw new Error('Current step is out of limits')
    }

    const contextValue = useMemo(() => ({ state, dispatch }), [state, dispatch])

    return <StepsContext.Provider value={contextValue}>{children}</StepsContext.Provider>
}

export const Step: React.FunctionComponent<React.PropsWithChildren<StepProps>> = ({ children, borderColor }) => {
    const { state } = useStepsContext()
    const { setCurrent, stepIndex } = useStepListContext()
    const { current, steps } = state

    // Marking all previous steps active helps when we using the debug option to start the flow in
    // the middle.
    const didSeeStep = steps[stepIndex]?.isVisited || stepIndex <= current

    const disabled = !didSeeStep
    const active = didSeeStep

    return (
        <li
            className={classNames(
                disabled && stepsStyles.disabled,
                stepsStyles.listItem,
                stepsStyles.active,
                borderColor && stepsStyles[`color${upperFirst(borderColor)}` as keyof typeof stepsStyles]
            )}
            aria-current={active}
        >
            <button
                type="button"
                tabIndex={active ? 0 : -1}
                disabled={disabled}
                className={stepsStyles.button}
                onClick={() => !disabled && setCurrent()}
            >
                {children}
            </button>
        </li>
    )
}

export const StepList: React.FunctionComponent<React.PropsWithChildren<StepListProps>> = ({
    children,
    numeric,
    className,
}) => {
    const { state, dispatch } = useStepsContext()

    const { initialStep } = state

    const childrenArray = React.Children.toArray(children)

    if (childrenArray.length !== state.totalSteps) {
        throw new Error('StepList must include as many steps as defined by totalSteps')
    }

    useEffect(() => {
        const steps = childrenArray.reduce((accumulator: StepsInterface, _current, index) => {
            const value = {
                index: index + 1,
                isFirstStep: index === 0,
                isLastStep: index === childrenArray.length - 1,
                isVisited: initialStep === index + 1,
                isComplete: false,
            }

            accumulator[index + 1] = value
            return accumulator
        }, {})

        dispatch({ type: 'SET_STEPS', payload: { steps } })
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [dispatch])

    const element = React.Children.map(children, (child: React.ReactElement<StepProps>, index) => {
        const setCurrent = (): void => {
            dispatch({ type: 'SET_CURRENT_STEP', payload: { index: index + 1 } })
        }

        return <StepListContext.Provider value={{ setCurrent, stepIndex: index + 1 }}>{child}</StepListContext.Provider>
    })

    return (
        <nav className={className}>
            {numeric ? (
                <ol className={stepsStyles.listNumeric}>{element}</ol>
            ) : (
                <ul className={stepsStyles.list}>{element}</ul>
            )}
        </nav>
    )
}

export const StepPanels: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({ children }) => {
    const { state } = useStepsContext()
    const { current } = state

    const childrenArray = React.Children.toArray(children)
    const indexArray = current - 1

    if (!children) {
        throw new Error('StepPanels must include at least one child')
    }

    if (childrenArray.length !== state.totalSteps) {
        throw new Error('StepPanels must include as many steps as defined by totalSteps')
    }

    if (indexArray < 0 || current > childrenArray.length) {
        throw new Error(
            'The step-index is out of the boundaries. Check if the number of steps and your initialStep or setStep assignation are in concordance.'
        )
    }

    return <div className="mt-4 pb-3">{childrenArray[indexArray]}</div>
}

export const StepPanel: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({ children }) => <>{children}</>
