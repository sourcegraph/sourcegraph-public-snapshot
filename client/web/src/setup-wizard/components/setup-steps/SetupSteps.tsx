import {
    useRef,
    Ref,
    createContext,
    FC,
    ReactNode,
    HTMLAttributes,
    useMemo,
    useContext,
    useCallback,
    useEffect,
} from 'react'

import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'
import { createPortal } from 'react-dom'
import { useLocation, useNavigate, Routes, Route, Navigate, matchPath } from 'react-router-dom-v5-compat'

import { Button, Icon } from '@sourcegraph/wildcard'

import styles from './SetupSteps.module.scss'

export interface StepConfiguration {
    id: string
    path: string
    name: string
    render: () => ReactNode
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

interface SetupStepURLContext {
    currentStep: StepConfiguration
    activeStepIndex: number
}

export const SetupStepsRoot: FC<SetupStepsProps> = props => {
    const { initialStepId, steps, onStepChange } = props

    const navigate = useNavigate()
    const location = useLocation()
    const nextButtonPortalRef = useRef<HTMLDivElement>(null)

    // Resolve current setup step and its index by URL matches
    const { currentStep, activeStepIndex } = useMemo<SetupStepURLContext>(() => {
        // Try to find step by URL based on available steps
        const urlStepIndex = steps.findIndex(step => matchPath(step.path, location.pathname) !== null)

        if (urlStepIndex !== -1) {
            return {
                activeStepIndex: urlStepIndex,
                currentStep: steps[urlStepIndex],
            }
        }

        // Try to find step by pre-saved settings if URL doesn't resolve any step
        const savedStepIndex = steps.findIndex(step => step.id === initialStepId)

        if (savedStepIndex !== -1) {
            return {
                activeStepIndex: savedStepIndex,
                currentStep: steps[savedStepIndex],
            }
        }

        // Fallback on the first available step if URL doesn't match any step, and we
        // don't have any pre-saved step
        return {
            activeStepIndex: 0,
            currentStep: steps[0],
        }
    }, [location, initialStepId, steps])

    useEffect(() => {
        onStepChange(currentStep)
    }, [currentStep, onStepChange])

    const handleGoToNextStep = useCallback(() => {
        const nextStepIndex = activeStepIndex + 1

        if (nextStepIndex < steps.length) {
            const nextStep = steps[nextStepIndex]

            navigate(nextStep.path)
        }
    }, [activeStepIndex, steps, navigate])

    const handleGoToPrevStep = useCallback(() => {
        const prevStepIndex = activeStepIndex - 1

        if (prevStepIndex >= 0) {
            const prevStep = steps[prevStepIndex]

            navigate(prevStep.path)
        }
    }, [activeStepIndex, steps, navigate])

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
                    <Routes>
                        {steps.map(step => (
                            <Route key="hardcoded-key" path={step.path} element={step.render()} />
                        ))}
                        <Route path="*" element={<Navigate to={currentStep.path} />} />
                    </Routes>
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
                    <small className={styles.headerStepLabel}>{step.name}</small>
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
