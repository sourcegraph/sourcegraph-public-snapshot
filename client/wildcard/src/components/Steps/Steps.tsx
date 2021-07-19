import classNames from 'classnames'
import { upperFirst } from 'lodash'
import React, { useEffect, useMemo, useReducer, useRef, MutableRefObject } from 'react'

import { StepsContext, useStepsContext, StepListContext, useStepListContext, Steps as StepsInterface } from './context'
import { initialState, reducer } from './reducer'
import stepsStyles from './Steps.module.scss'

type Color = 'orange' | 'blue' | 'purple'

export interface StepProps {
    borderColor: Color
    children: React.ReactNode
}

interface StepListProps {
    numeric: boolean
    children: React.ReactElement<StepProps> | React.ReactElement<StepProps, string | React.JSXElementConstructor<any>>[]
}

export interface StepsProps {
    children: React.ReactElement<StepProps> | React.ReactElement<StepProps>[]
    initialStep: number
}

export const Steps: React.FunctionComponent<StepsProps> = ({ initialStep = 1, children }) => {
    const [state, dispatch] = useReducer(reducer, initialState(initialStep))

    if (!children) {
        throw new Error('Steps must include at least one child')
    }

    if (initialStep < 1 || initialStep > React.Children.count(children)) {
        console.warn('current step is out of limits')
    }

    const contextValue = useMemo(() => ({ state, dispatch }), [state, dispatch])

    return <StepsContext.Provider value={contextValue}>{children}</StepsContext.Provider>
}

export const Step: React.FunctionComponent<StepProps> = ({ children, borderColor }) => {
    const { state } = useStepsContext()
    const { setCurrent, stepIndex } = useStepListContext()

    const { current, steps } = state
    const disabled = !steps[stepIndex]?.isVisited && current !== steps[stepIndex]?.index
    const active = current === steps[stepIndex]?.index || steps[stepIndex]?.isVisited

    return (
        <li
            role="presentation"
            className={classNames(
                stepsStyles.cursorPointer,
                disabled && stepsStyles.disabled,
                stepsStyles.listItem,
                stepsStyles.active,
                borderColor && stepsStyles[`color${upperFirst(borderColor)}` as keyof typeof stepsStyles]
            )}
            aria-current={active}
            onClick={() => !disabled && setCurrent()}
        >
            {children}
        </li>
    )
}

export const StepList: React.FunctionComponent<StepListProps> = ({ children, numeric }) => {
    const { state, dispatch } = useStepsContext()

    const { initialStep } = state

    const childrenArray = React.Children.toArray(children)

    const stepsCollection: MutableRefObject<() => StepsInterface> = useRef(() =>
        childrenArray.reduce((accumulator: StepsInterface, _current, index) => {
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
    )

    useEffect(() => {
        dispatch({ type: 'SET_STEPS', payload: { steps: stepsCollection.current() } })
    }, [dispatch, stepsCollection])

    const element = React.Children.map(children, (child: React.ReactElement<StepProps>, index) => {
        if (child.type !== Step) {
            throw new Error(`${child.type.toString()} element is not <Step> component`)
        }

        const setCurrent = (): void => {
            dispatch({ type: 'SET_CURRENT_STEP', payload: { index: index + 1 } })
        }

        return <StepListContext.Provider value={{ setCurrent, stepIndex: index + 1 }}>{child}</StepListContext.Provider>
    })

    return (
        <nav className={stepsStyles.stepsWrapper} aria-label="progress">
            {numeric ? <ol>{element}</ol> : <ul>{element}</ul>}
        </nav>
    )
}

export const StepPanels: React.FunctionComponent = ({ children }) => {
    const { state } = useStepsContext()
    const { current } = state

    const childrenArray = React.Children.toArray(children)

    if (!children) {
        throw new Error('You need to add the same number of <StepPanels> and <Step> Components')
    }

    return <div className="mt-4 pb-3">{childrenArray[current - 1]}</div>
}

export const StepPanel: React.FunctionComponent = ({ children }) => <>{children}</>

export const StepActions: React.FunctionComponent = ({ children }) => <>{children}</>
