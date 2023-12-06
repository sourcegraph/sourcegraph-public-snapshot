import { type FC, useCallback, useLayoutEffect, useEffect } from 'react'

import { gql } from '@apollo/client'
import { readTextFile, BaseDirectory } from '@tauri-apps/api/fs'
import { appWindow, LogicalSize } from '@tauri-apps/api/window'
import { useNavigate } from 'react-router-dom'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Theme, ThemeContext, ThemeSetting, useTheme } from '@sourcegraph/shared/src/theme'

import { getWebGraphQLClient } from '../../../backend/graphql'
import { PageTitle } from '../../../components/PageTitle'
import type { AppUserConnectDotComAccountResult } from '../../../graphql-operations'
import { SetupStepsContent, SetupStepsRoot, type StepConfiguration } from '../../../setup-wizard'
import { useLocalExternalServices } from '../../../setup-wizard/components'
import { FooterWidgetPortal } from '../../../setup-wizard/components/setup-steps'
import { saveAccessToken } from '../AppAuthCallbackPage'

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
export const AppSetupWizard: FC<TelemetryProps & TelemetryV2Props> = ({ telemetryService, telemetryRecorder }) => {
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

    const { services: presentServices, addRepositories, loaded } = useLocalExternalServices()
    const navigate = useNavigate()

    useEffect(() => {
        if (loaded) {
            // ~/Library/Application Support/com.sourcegraph.cody/vscode.json
            readTextFile('vscode.json', { dir: BaseDirectory.AppData })
                .then(async data => {
                    const { dotcomAccessToken = '', repoPaths = [] } = JSON.parse(data) as {
                        dotcomAccessToken?: string
                        repoPaths?: string[]
                    }

                    if (!(await isSourcegraphAccountConnected()) && !dotcomAccessToken) {
                        return
                    }

                    await saveAccessToken(dotcomAccessToken)

                    if (!(await isSourcegraphAccountConnected())) {
                        return
                    }

                    const missingRepoPaths = repoPaths.filter(
                        path => !presentServices.find(service => service.path === path)
                    )

                    if (missingRepoPaths.length > 0) {
                        await addRepositories(missingRepoPaths)
                    }

                    setStepId('local-repositories')
                    navigate('/app-setup/local-repositories?from=vscode')
                })
                .catch(() => {})
        }
        // NOTE(naman): we are updating the `presentServices` state in the above function
        // and we want to run this effect only once when the `loaded` state is updated.

        /* eslint-disable-next-line react-hooks/exhaustive-deps */
    }, [loaded])

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
                        telemetryRecorder={telemetryRecorder}
                        className={styles.content}
                        isCodyApp={true}
                        setStepId={setStepId}
                    />

                    <FooterWidgetPortal className={styles.footer} />
                </SetupStepsRoot>
            </div>
        </ThemeContext.Provider>
    )
}

const IS_SOURCEGRAPH_ACCOUNT_CONNECTED_QUERY = gql`
    query AppUserConnectDotComAccount {
        site {
            id
            appHasConnectedDotComAccount
        }
    }
`

const isSourcegraphAccountConnected = async (): Promise<boolean> => {
    const client = await getWebGraphQLClient()
    const { data } = await client.query<AppUserConnectDotComAccountResult>({
        query: IS_SOURCEGRAPH_ACCOUNT_CONNECTED_QUERY,
    })

    return data?.site?.appHasConnectedDotComAccount ?? false
}
