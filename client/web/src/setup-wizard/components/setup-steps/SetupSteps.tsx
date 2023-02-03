import {
    useRef,
    Ref,
    createContext,
    useState,
    FC,
    ReactElement,
    ReactNode,
    HTMLAttributes,
    useMemo,
    useContext,
    useCallback,
} from 'react'

import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'
import { createPortal } from 'react-dom'
import { Switch, Redirect, Route } from 'react-router'
import { useNavigate } from 'react-router-dom-v5-compat'

import { Button, Icon } from '@sourcegraph/wildcard'

import styles from './SetupSteps.module.scss'

export interface StepConfiguration {
    id: string
    path: string
    name: string
    render: ReactElement | (() => ReactNode)
}

interface SetupStepsContextData {
    steps: StepConfiguration[]
    nextButtonPortalElement: HTMLDivElement | null
    onNextStep: () => void
}

const SetupStepsContext = createContext<SetupStepsContextData>({
    steps: [],
    nextButtonPortalElement: null,
    onNextStep: () => {},
})

interface SetupStepsProps {
    initialStepId: string | undefined
    steps: StepConfiguration[]
    onStepChange: (nextStep: StepConfiguration) => void
}

export const SetupStepsRoot: FC<SetupStepsProps> = props => {
    const { initialStepId, steps, onStepChange } = props

    const navigate = useNavigate()
    const [stepId, setStepId] = useState(initialStepId)
    const nextButtonPortalRef = useRef<HTMLDivElement>(null)

    const rawActiveStep = steps.findIndex(step => step.id === stepId)
    const activeStepIndex = rawActiveStep !== -1 ? rawActiveStep : 0
    const availableSteps = steps.filter((step, index) => index <= activeStepIndex)

    const handleGoToNextStep = useCallback(() => {
        const nextStepIndex = activeStepIndex + 1

        if (nextStepIndex < steps.length) {
            const nextStep = steps[nextStepIndex]

            setStepId(nextStep.id)
            navigate(nextStep.path)
            onStepChange(nextStep)
        }
    }, [activeStepIndex, navigate, onStepChange, steps])

    const handleGoToPrevStep = (): void => {
        const prevStepIndex = activeStepIndex - 1

        if (prevStepIndex >= 0) {
            const prevStep = steps[prevStepIndex]

            setStepId(prevStep.id)
            navigate(prevStep.path)
            onStepChange(prevStep)
        }
    }

    const cachedContext = useMemo(
        () => ({
            steps,
            nextButtonPortalElement: nextButtonPortalRef.current,
            onNextStep: handleGoToNextStep,
        }),
        [handleGoToNextStep, steps]
    )

    return (
        <SetupStepsContext.Provider value={cachedContext}>
            <div className={styles.root}>
                <SetupStepsHeader steps={steps} activeStepIndex={activeStepIndex} />
                <div className={styles.content}>
                    <Switch>
                        {availableSteps.map(step => (
                            <Route key={step.id} path={step.path} exact={true}>
                                {step.render}
                            </Route>
                        ))}

                        <Redirect to={steps[activeStepIndex].path} />
                    </Switch>
                </div>
                <SetupStepsFooter
                    steps={steps}
                    activeStepIndex={activeStepIndex}
                    nextButtonPortalRef={nextButtonPortalRef}
                    onPrevStep={handleGoToPrevStep}
                    onNextStep={handleGoToNextStep}
                />
            </div>
        </SetupStepsContext.Provider>
    )
}

interface SetupStepsHeaderProps extends HTMLAttributes<HTMLElement> {
    steps: StepConfiguration[]
    activeStepIndex: number
}

export const SetupStepsHeader: FC<SetupStepsHeaderProps> = props => {
    const { steps, activeStepIndex, className, ...attributes } = props

    return (
        <header {...attributes} className={classNames(className, styles.header)}>
            {steps.map((step, index) => (
                <div key={step.id} className={styles.headerStep}>
                    <span
                        className={classNames(styles.headerStepNumber, {
                            [styles.headerStepNumberCompleted]: index < activeStepIndex,
                            [styles.headerStepNumberDisabled]: index > activeStepIndex,
                        })}
                    >
                        {index + 1}
                    </span>
                    <span className={styles.headerStepLabel}>{step.name}</span>
                </div>
            ))}
        </header>
    )
}

interface SetupStepsFooterProps {
    steps: StepConfiguration[]
    activeStepIndex: number
    nextButtonPortalRef: Ref<HTMLDivElement>
    onPrevStep: () => void
    onNextStep: () => void
}

export const SetupStepsFooter: FC<SetupStepsFooterProps> = props => {
    const { steps, activeStepIndex, nextButtonPortalRef, onPrevStep, onNextStep } = props

    return (
        <footer className={styles.navigation}>
            <div className={styles.navigationInner}>
                {activeStepIndex > 0 && (
                    <Button variant="secondary" onClick={onPrevStep}>
                        <Icon svgPath={mdiChevronLeft} aria-hidden={true} /> Go to previous step
                    </Button>
                )}

                <div ref={nextButtonPortalRef} className={styles.navigationNextPortal} />
                <Button variant="primary" className={styles.navigationNext} onClick={onNextStep}>
                    {activeStepIndex < steps.length - 1 ? 'Next' : 'Finish'}{' '}
                    <Icon svgPath={mdiChevronRight} aria-hidden={true} />
                </Button>
            </div>
        </footer>
    )
}

interface CustomNextButtonProps {
    label: string
    disabled: boolean
}

export const CustomNextButton: FC<CustomNextButtonProps> = props => {
    const { label, disabled } = props
    const { nextButtonPortalElement, onNextStep } = useContext(SetupStepsContext)

    if (!nextButtonPortalElement) {
        return null
    }

    return createPortal(
        <Button variant="primary" disabled={disabled} onClick={onNextStep}>
            {label}
        </Button>,
        nextButtonPortalElement
    )
}
