import { FC } from 'react'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { Container, H1, H2 } from '@sourcegraph/wildcard'

import { BrandLogo } from '../components/branding/BrandLogo'

import { SetupStepsRoot, CustomNextButton, StepConfiguration } from './components/setup-steps'

import styles from './Setup.module.scss'

const SETUP_STEPS: StepConfiguration[] = [
    {
        id: '001',
        name: 'Add local repositories',
        path: '/setup/local-repositories',
        render: () => <H2>Hello local repositories step</H2>,
    },
    {
        id: '002',
        name: 'Add remote repositories',
        path: '/setup/remote-repositories',
        render: () => (
            <Container>
                <H2>Hello remote repositories step</H2>
                <CustomNextButton label="Custom next step label" disabled={true} />
            </Container>
        ),
    },
    {
        id: '003',
        name: 'Sync repositories',
        path: '/setup/sync-repositories',
        render: () => <H2>Hello sync repositories step</H2>,
    },
]

export const SetupWizard: FC = props => {
    const [activeStepId, setStepId, status] = useTemporarySetting('setup.activeStepId')

    if (status !== 'loaded') {
        return null
    }

    const handleStepChange = (step: StepConfiguration): void => {
        setStepId(step.id)
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
