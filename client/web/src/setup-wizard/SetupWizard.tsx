import { FC, ReactElement, useCallback, useMemo } from 'react'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { H1, H2, Text } from '@sourcegraph/wildcard'

import { BrandLogo } from '../components/branding/BrandLogo'
import { PageTitle } from '../components/PageTitle'
import { SiteAdminRepositoriesContainer } from '../site-admin/SiteAdminRepositoriesContainer'

import { RemoteRepositoriesStep } from './components/remote-repositories-step'
import { SetupStepsRoot, SetupStepsContent, SetupStepsFooter, StepConfiguration } from './components/setup-steps'

import styles from './Setup.module.scss'

const CORE_STEPS: StepConfiguration[] = [
    {
        id: 'remote-repositoires',
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

interface SetupWizardProps {
    isSourcegraphApp: boolean
}

export const SetupWizard: FC<SetupWizardProps> = props => {
    const { isSourcegraphApp } = props

    const [activeStepId, setStepId, status] = useTemporarySetting('setup.activeStepId')
    const steps = useMemo(() => (isSourcegraphApp ? SOURCEGRAPH_APP_STEPS : CORE_STEPS), [isSourcegraphApp])

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
            <PageTitle title="Setup" />
            <SetupStepsRoot initialStepId={activeStepId} steps={steps} onStepChange={handleStepChange}>
                <div className={styles.content}>
                    <header className={styles.header}>
                        <BrandLogo variant="logo" isLightTheme={false} className={styles.logo} />

                        <H2 as={H1} className="font-weight-normal text-white mt-3 mb-4">
                            Welcome to Sourcegraph! Let's get started.
                        </H2>
                    </header>

                    <SetupStepsContent />
                </div>

                <SetupStepsFooter className={styles.footer} />
            </SetupStepsRoot>
        </div>
    )
}

function LocalRepositoriesStep(props: any): ReactElement {
    return <H2 {...props}>Hello local repositories step</H2>
}

function SyncRepositoriesStep(props: any): ReactElement {
    return (
        <section {...props}>
            <Text className="mb-2">
                It may take a few moments to clone and index each repository. Repository statuses are displayed below.
            </Text>
            <SiteAdminRepositoriesContainer />
        </section>
    )
}
