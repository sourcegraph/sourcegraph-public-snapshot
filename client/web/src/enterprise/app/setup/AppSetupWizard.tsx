import { FC, useCallback } from 'react'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { PageTitle } from '../../../components/PageTitle'
import { StepConfiguration, SetupStepsContent, SetupStepsRoot } from '../../../setup-wizard'

import { AppWelcomeSetupPage, AddLocalRepositoriesSetupPage, InstallExtensionsSetupPage } from './AppSetupSteps'

import styles from './AppSetupWizard.module.scss'

const APP_SETUP_STEPS: StepConfiguration[] = [
    {
        id: 'welcome',
        name: 'Welcome page',
        path: 'welcome',
        component: AppWelcomeSetupPage,
    },
    {
        id: 'local-repositories',
        name: 'Add local repositories',
        path: 'local-repositories',
        component: AddLocalRepositoriesSetupPage,
    },
    {
        id: 'install-extensions',
        name: 'Install Sourcegraph extensions',
        path: 'install-extensions',
        component: InstallExtensionsSetupPage,
        nextURL: '/',
        onNext: () => {
            localStorage.setItem('app.setup.finished', 'true')
        },
    },
]

/**
 * App related setup wizard component, see {@link SetupWizard} component
 * for any other deploy type setup flows.
 */
export const AppSetupWizard: FC<TelemetryProps> = ({ telemetryService }) => {
    const [activeStepId, setStepId, status] = useTemporarySetting('app-setup.activeStepId')

    const handleStepChange = useCallback(
        (nextStep: StepConfiguration): void => {
            const currentStepIndex = APP_SETUP_STEPS.findIndex(step => step.id === nextStep.id)
            const isLastStep = currentStepIndex === APP_SETUP_STEPS.length - 1

            // Reset the last visited step if you're on the last step in the
            // setup pipeline
            setStepId(!isLastStep ? nextStep.id : '')
        },
        [setStepId]
    )

    if (status !== 'loaded') {
        return null
    }

    return (
        <div className={styles.root}>
            <PageTitle title="Sourcegraph App setup" />

            <SetupStepsRoot
                baseURL="/app-setup/"
                initialStepId={activeStepId}
                steps={APP_SETUP_STEPS}
                onStepChange={handleStepChange}
            >
                <SetupStepsContent telemetryService={telemetryService} className={styles.content} />
            </SetupStepsRoot>
        </div>
    )
}
