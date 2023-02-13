import { FC, ReactElement, useCallback } from 'react'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { H1, H2 } from '@sourcegraph/wildcard'

import { BrandLogo } from '../components/branding/BrandLogo'

import { RemoteRepositoriesStep } from './components/remote-repositories-step'
import { SetupStepsRoot, StepConfiguration } from './components/setup-steps'

import styles from './Setup.module.scss'

const SETUP_STEPS: StepConfiguration[] = [
    {
        id: '001',
        name: 'Add local repositories',
        path: '/setup/local-repositories',
        component: LocalRepositoriesStep,
    },
    {
        id: '002',
        name: 'Add remote repositories',
        path: '/setup/remote-repositories',
        component: RemoteRepositoriesStep,
    },
    {
        id: '003',
        name: 'Sync repositories',
        path: '/setup/sync-repositories',
        component: () => <H2>Hello sync repositories step</H2>,
    },
]

export const SetupWizard: FC = props => {
    const [activeStepId, setStepId, status] = useTemporarySetting('setup.activeStepId')

    const handleStepChange = useCallback(
        (step: StepConfiguration): void => {
            setStepId(step.id)
        },
        [setStepId]
    )

    if (status !== 'loaded') {
        return null
    }

    return (
        <div className={styles.root}>
            <header className={styles.header}>
                <BrandLogo variant="logo" isLightTheme={false} className={styles.logo} />

                <H2 as={H1} className="font-weight-normal text-white mt-3 mb-4">
                    Welcome to Sourcegraph! Let's get started.
                </H2>
            </header>

            <SetupStepsRoot initialStepId={activeStepId} steps={SETUP_STEPS} onStepChange={handleStepChange} />
        </div>
    )
}

function LocalRepositoriesStep(props: any): ReactElement {
    return <H2 {...props}>Hello local repositories step</H2>
}
