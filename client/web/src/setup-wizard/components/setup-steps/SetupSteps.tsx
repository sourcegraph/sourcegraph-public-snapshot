import {
    createContext,
    ComponentType,
    FC,
    HTMLAttributes,
    useMemo,
    useContext,
    useCallback,
    useEffect,
    useState,
    ReactNode,
    PropsWithChildren,
} from 'react'

import { ApolloClient, useApolloClient } from '@apollo/client'
import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'
import { noop } from 'lodash'
import { createPortal } from 'react-dom'
import { useLocation, useNavigate, Routes, Route, Navigate, matchPath } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import styles from './SetupSteps.module.scss'

interface StepComponentProps extends TelemetryProps {
    className?: string
}

export interface StepConfiguration {
    id: string
    path: string
    name: string
    nextURL?: string
    component: ComponentType<StepComponentProps>
    onNext?: (client: ApolloClient<{}>) => void
}

interface SetupStepsContextData {
    steps: StepConfiguration[]
    activeStepIndex: number
    footerPortal: HTMLDivElement | null
    nextButtonPortal: HTMLDivElement | null
    setFooterPortal: (container: HTMLDivElement | null) => void
    setNextButtonPortal: (container: HTMLDivElement | null) => void
    onSkip: () => void
    onPrevStep: () => void
    onNextStep: () => void
}

const SetupStepsContext = createContext<SetupStepsContextData>({
    steps: [],
    activeStepIndex: 0,
    footerPortal: null,
    nextButtonPortal: null,
    setFooterPortal: noop,
    setNextButtonPortal: noop,
    onSkip: noop,
    onPrevStep: noop,
    onNextStep: noop,
})

interface SetupStepsProps {
    initialStepId: string | undefined
    steps: StepConfiguration[]
    children?: ReactNode
    onSkip: () => void
    onStepChange: (nextStep: StepConfiguration) => void
}

interface SetupStepURLContext {
    activeStepIndex: number
}

export const SetupStepsRoot: FC<SetupStepsProps> = props => {
    const { initialStepId, steps, children, onSkip, onStepChange } = props

    const navigate = useNavigate()
    const location = useLocation()
    const client = useApolloClient()

    const [nextButtonPortal, setNextButtonPortal] = useState<HTMLDivElement | null>(null)
    const [footerPortal, setFooterPortal] = useState<HTMLDivElement | null>(null)

    // Resolve current setup step and its index by URL matches
    const { activeStepIndex } = useMemo<SetupStepURLContext>(() => {
        // Try to find step by URL based on available steps
        const urlStepIndex = steps.findIndex(step => matchPath(step.path, location.pathname) !== null)

        if (urlStepIndex !== -1) {
            return { activeStepIndex: urlStepIndex }
        }

        // Try to find step by pre-saved settings if URL doesn't resolve any step
        const savedStepIndex = steps.findIndex(step => step.id === initialStepId)

        if (savedStepIndex !== -1) {
            return { activeStepIndex: savedStepIndex }
        }

        // Fallback on the first available step if URL doesn't match any step, and we
        // don't have any pre-saved step
        return { activeStepIndex: 0 }
    }, [location, initialStepId, steps])

    const currentStep = steps[activeStepIndex]

    useEffect(() => {
        onStepChange(currentStep)
    }, [currentStep, onStepChange])

    const handleGoToNextStep = useCallback(() => {
        const activeStep = steps[activeStepIndex]
        const nextStepIndex = activeStepIndex + 1

        activeStep.onNext?.(client)

        if (activeStep.nextURL) {
            navigate(activeStep.nextURL)
            return
        }

        if (nextStepIndex < steps.length) {
            const nextStep = steps[nextStepIndex]

            navigate(nextStep.path)
        }
    }, [activeStepIndex, steps, navigate, client])

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
            activeStepIndex,
            footerPortal,
            nextButtonPortal,
            setFooterPortal,
            setNextButtonPortal,
            onSkip,
            onPrevStep: handleGoToPrevStep,
            onNextStep: handleGoToNextStep,
        }),
        [steps, activeStepIndex, footerPortal, nextButtonPortal, onSkip, handleGoToPrevStep, handleGoToNextStep]
    )

    return <SetupStepsContext.Provider value={cachedContext}>{children}</SetupStepsContext.Provider>
}

interface SetupStepsContentProps extends TelemetryProps, HTMLAttributes<HTMLElement> {}

export const SetupStepsContent: FC<SetupStepsContentProps> = props => {
    const { className, telemetryService, ...attributes } = props
    const { steps, activeStepIndex } = useContext(SetupStepsContext)

    return (
        <div {...attributes} className={classNames(styles.root, className)}>
            <SetupStepsHeader steps={steps} activeStepIndex={activeStepIndex} />
            <Routes>
                {steps.map(({ path, component: Component }) => (
                    <Route
                        key="hardcoded-key"
                        path={`${path}/*`}
                        element={<Component className={styles.content} telemetryService={telemetryService} />}
                    />
                ))}
                <Route path="*" element={<Navigate to={steps[activeStepIndex].path} replace={true} />} />
            </Routes>
        </div>
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
                    <small
                        data-label-text={step.name}
                        className={classNames(styles.headerStepLabel, {
                            [styles.headerStepLabelActive]: index === activeStepIndex,
                        })}
                    >
                        {step.name}
                    </small>
                </div>
            ))}
        </header>
    )
}

export const SetupStepsFooter: FC<HTMLAttributes<HTMLElement>> = props => {
    const { className, ...attributes } = props

    const { steps, activeStepIndex, setNextButtonPortal, setFooterPortal, onSkip, onPrevStep, onNextStep } =
        useContext(SetupStepsContext)

    return (
        <footer {...attributes} className={classNames(styles.footer, className)}>
            <div ref={setFooterPortal} className={styles.footerWidget} />
            <div className={styles.footerNavigation}>
                <div className={styles.footerInnerNavigation}>
                    <Button variant="link" className={styles.footerSkip} onClick={onSkip}>
                        Skip setup
                    </Button>

                    {activeStepIndex > 0 && (
                        <Button variant="secondary" onClick={onPrevStep}>
                            <Icon svgPath={mdiChevronLeft} aria-hidden={true} /> Previous
                        </Button>
                    )}

                    <div ref={setNextButtonPortal} className={styles.footerNextPortal} />
                    <Button variant="primary" className={styles.footerNext} onClick={onNextStep}>
                        {activeStepIndex < steps.length - 1 ? 'Next' : 'Finish'}{' '}
                        <Icon svgPath={mdiChevronRight} aria-hidden={true} />
                    </Button>
                </div>
            </div>
        </footer>
    )
}

export const FooterWidget: FC<PropsWithChildren<{}>> = props => {
    const { children } = props
    const { footerPortal } = useContext(SetupStepsContext)

    if (!footerPortal) {
        return null
    }

    return createPortal(children, footerPortal)
}

interface CustomNextButtonProps {
    label: string
    disabled?: boolean
    tooltip?: string
    onClick?: () => void
}

export const CustomNextButton: FC<CustomNextButtonProps> = props => {
    const { label, disabled, tooltip, onClick } = props
    const { nextButtonPortal, onNextStep } = useContext(SetupStepsContext)

    if (!nextButtonPortal) {
        return null
    }

    const handleNextClick = (): void => {
        onClick?.()
        onNextStep()
    }

    return createPortal(
        <Tooltip content={tooltip}>
            <Button variant="primary" disabled={disabled} onClick={handleNextClick}>
                {label}
                <Icon svgPath={mdiChevronRight} aria-hidden={true} />
            </Button>
        </Tooltip>,
        nextButtonPortal
    )
}
