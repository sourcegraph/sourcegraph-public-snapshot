import { type FC, useCallback } from 'react'

import type { ApolloClient } from '@apollo/client'
import { useNavigate } from 'react-router-dom'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { H1, H2, useLocalStorage } from '@sourcegraph/wildcard'

import { BrandLogo } from '../components/branding/BrandLogo'
import { PageTitle } from '../components/PageTitle'
import { refreshSiteFlags } from '../site/backend'

import {
    type StepConfiguration,
    SetupStepsRoot,
    SetupStepsHeader,
    SetupStepsContent,
    SetupStepsFooter,
    RemoteRepositoriesStep,
    SyncRepositoriesStep,
} from './components'

import styles from './Setup.module.scss'

const CORE_STEPS: StepConfiguration[] = [
    {
        id: 'remote-repositories',
        name: 'Add remote repositories',
        path: '/setup/remote-repositories',
        component: RemoteRepositoriesStep,
        // If user clicked next button in setup remote repositories
        // this mean that setup was completed, and they're ready to go
        // to app UI. See https://github.com/sourcegraph/sourcegraph/issues/50122
        onNext: (client: ApolloClient<{}>) => {
            // Mutate initial needsRepositoryConfiguration value
            // in order to avoid loop in Layout page redirection logic
            // TODO Remove this as soon as we have a proper Sourcegraph context store
            window.context.needsRepositoryConfiguration = false

            // Update global site flags in order to fix global navigation items about
            // setup instance state
            refreshSiteFlags(client).then(
                () => {},
                () => {}
            )
        },
    },
    {
        id: 'sync-repositories',
        name: 'Sync repositories',
        path: '/setup/sync-repositories',
        nextURL: '/search',
        component: SyncRepositoriesStep,
    },
]

interface SetupWizardProps extends TelemetryProps {}

export const SetupWizard: FC<SetupWizardProps> = props => {
    const { telemetryService } = props

    const navigate = useNavigate()
    const [activeStepId, setStepId, status] = useTemporarySetting('setup.activeStepId')

    // We use local storage since async nature of temporal settings doesn't allow us to
    // use it for wizard redirection logic (see layout component there we read this state
    // about the setup wizard availability and redirect to the wizard if it wasn't skipped already.
    // eslint-disable-next-line no-restricted-syntax
    const [, setSkipWizardState] = useLocalStorage('setup.skipped', false)
    const steps = CORE_STEPS

    const handleStepChange = useCallback(
        (nextStep: StepConfiguration): void => {
            const currentStepIndex = steps.findIndex(step => step.id === nextStep.id)
            const isLastStep = currentStepIndex === steps.length - 1

            // Reset the last visited step if you're on the last step in the
            // setup pipeline
            setStepId(!isLastStep ? nextStep.id : '')
        },
        [setStepId, steps]
    )

    const handleSkip = useCallback(() => {
        setSkipWizardState(true)
        telemetryService.log('SetupWizardQuits')
        navigate('/search')
    }, [navigate, telemetryService, setSkipWizardState])

    if (status !== 'loaded') {
        return null
    }

    return (
        <div className={styles.root}>
            <PageTitle title="Setup" />
            <SetupStepsRoot
                initialStepId={activeStepId}
                steps={steps}
                onSkip={handleSkip}
                onStepChange={handleStepChange}
            >
                <div className={styles.content}>
                    <header className={styles.header}>
                        <BrandLogo variant="logo" isLightTheme={false} className={styles.logo} />

                        <H2 as={H1} className="font-weight-normal text-white mt-3 mb-4">
                            Welcome to Sourcegraph! Let's get started.
                        </H2>
                    </header>

                    <SetupStepsHeader className={styles.steps} />
                    <SetupStepsContent
                        contentContainerClass={styles.contentContainer}
                        telemetryService={telemetryService}
                    />
                </div>

                <SetupStepsFooter className={styles.footer} />
            </SetupStepsRoot>
        </div>
    )
}
