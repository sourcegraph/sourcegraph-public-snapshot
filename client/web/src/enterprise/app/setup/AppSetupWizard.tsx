import { type FC, useCallback, useLayoutEffect } from 'react'

import { appWindow, LogicalSize } from '@tauri-apps/api/window'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Theme, ThemeContext, ThemeSetting, useTheme } from '@sourcegraph/shared/src/theme'

import { PageTitle } from '../../../components/PageTitle'
import { SetupStepsContent, SetupStepsRoot, type StepConfiguration } from '../../../setup-wizard'
import { FooterWidgetPortal } from '../../../setup-wizard/components/setup-steps'

import { AppAllSetSetupStep } from './steps/AppAllSetSetupStep'
import { AppInstallExtensionsSetupStep } from './steps/AppInstallExtensionsSetupStep'
import { AddLocalRepositoriesSetupPage } from './steps/AppLocalRepositoriesSetupStep'
import { AppWelcomeSetupStep } from './steps/AppWelcomeSetupStep'

import styles from './AppSetupWizard.module.scss'

const APP_SETUP_STEPS: StepConfiguration[] = [
    {
        id: 'welcome',
        name: 'Welcome page',
        path: 'welcome',
        component: AppWelcomeSetupStep,
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
        component: AppInstallExtensionsSetupStep,
    },
    {
        id: 'all-set',
        name: 'All set',
        path: 'all-set',
        component: AppAllSetSetupStep,
        nextURL: '/',
        onNext: async () => {
            await appWindow.setResizable(true)
            await appWindow.setSize(new LogicalSize(1024, 768))
        },
    },
]

/**
 * App related setup wizard component, see {@link SetupWizard} component
 * for any other deploy type setup flows.
 */
export const AppSetupWizard: FC<TelemetryProps> = ({ telemetryService }) => {
    const { theme } = useTheme()
    const [activeStepId, setStepId, status] = useTemporarySetting('app-setup.activeStepId')

    const handleStepChange = useCallback(
        (nextStep: StepConfiguration): void => {
            setStepId(nextStep.id)
        },
        [setStepId]
    )

    // Force light theme when app setup wizard is rendered and return
    // original theme-based theme value when app wizard redirects to the
    // main app
    useLayoutEffect(() => {
        document.documentElement.classList.toggle('theme-light', true)
        document.documentElement.classList.toggle('theme-dark', false)

        return () => {
            const isLightTheme = theme === Theme.Light

            document.documentElement.classList.toggle('theme-light', isLightTheme)
            document.documentElement.classList.toggle('theme-dark', !isLightTheme)
        }
    }, [theme])

    // Local screen size when we are in app setup flow,
    // see last step onNext for window size restore and resize
    // unblock
    useLayoutEffect(() => {
        async function lockSize(): Promise<void> {
            await appWindow.setSize(new LogicalSize(940, 640))
            await appWindow.setResizable(false)
        }

        lockSize().catch(() => {})
    }, [])

    if (status !== 'loaded') {
        return null
    }

    return (
        <ThemeContext.Provider value={{ themeSetting: ThemeSetting.Light }}>
            <PageTitle title="Cody App setup" />

            <div className={styles.root}>
                <SetupStepsRoot
                    baseURL="/app-setup/"
                    initialStepId={activeStepId}
                    steps={APP_SETUP_STEPS}
                    onStepChange={handleStepChange}
                >
                    <SetupStepsContent
                        telemetryService={telemetryService}
                        className={styles.content}
                        isCodyApp={true}
                    />

                    <FooterWidgetPortal className={styles.footer} />
                </SetupStepsRoot>
            </div>
        </ThemeContext.Provider>
    )
}
