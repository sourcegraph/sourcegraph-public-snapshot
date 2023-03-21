import { FC, useCallback, useMemo } from 'react'

import { ApolloClient } from '@apollo/client'
import { useNavigate } from 'react-router-dom'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { H1, H2, useLocalStorage } from '@sourcegraph/wildcard'

import { BrandLogo } from '../components/branding/BrandLogo'
import { PageTitle } from '../components/PageTitle'
import { refreshSiteFlags } from '../site/backend'

import { LocalRepositoriesStep } from './components/local-repositories-step'
import { RemoteRepositoriesStep } from './components/remote-repositories-step'
import { SetupStepsRoot, SetupStepsContent, SetupStepsFooter, StepConfiguration } from './components/setup-steps'
import { SyncRepositoriesStep } from './components/SyncRepositoriesStep'

import styles from './Setup.module.scss'

const CORE_STEPS: StepConfiguration[] = [
    {
        id: 'remote-repositories',
        name: 'Add remote repositories',
        path: '/setup/remote-repositories',
        component: RemoteRepositoriesStep,
    },
    {
        id: 'sync-repositories',
        name: 'Sync repositories',
        path: '/setup/sync-repositories',
        nextURL: '/search',
        component: SyncRepositoriesStep,
        onNext: (client: ApolloClient<{}>) => {
            // Mutate initial needsRepositoryConfiguration value
            // in order to avoid loop in redirection logic
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
]

const SOURCEGRAPH_APP_STEPS = [
    {
        id: 'local-repositories',
        name: 'Add local repositories',
        path: '/setup/local-repositories',
        component: LocalRepositoriesStep,
    },
    ...CORE_STEPS,
]

interface SetupWizardProps extends TelemetryProps {
    isSourcegraphApp: boolean
}

export const SetupWizard: FC<SetupWizardProps> = props => {
    const { isSourcegraphApp, telemetryService } = props

    const navigate = useNavigate()
    const [activeStepId, setStepId, status] = useTemporarySetting('setup.activeStepId')

    // We use local storage since async nature of temporal settings doesn't allow us to
    // use it for wizard redirection logic (see layout component there we read this state
    // about the setup wizard availability and redirect to the wizard if it wasn't skipped already.
    // eslint-disable-next-line no-restricted-syntax
    const [, setSkipWizardState] = useLocalStorage('setup.skipped', false)
    const steps = useMemo(() => (isSourcegraphApp ? SOURCEGRAPH_APP_STEPS : CORE_STEPS), [isSourcegraphApp])

    const handleStepChange = useCallback(
        (step: StepConfiguration): void => {
            setStepId(step.id)
        },
        [setStepId]
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

                    <SetupStepsContent telemetryService={telemetryService} />
                </div>

                <SetupStepsFooter className={styles.footer} />
            </SetupStepsRoot>
        </div>
    )
}
