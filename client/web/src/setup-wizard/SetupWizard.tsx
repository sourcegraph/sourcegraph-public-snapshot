import { FC, useState } from 'react'

import { Text } from '@sourcegraph/wildcard'

import { BrandLogo } from '../components/branding/BrandLogo'
import { ThemePreference, useTheme } from '../theme'

import { SetupTabs, SetupList, SetupTab, SetupSteps } from './components/SetupTabs'
import { AddRepositoriesStep } from './steps/AddRepositoriesStep'
import { CloningRepositoriesStep } from './steps/CloningRepositoriesStep'
import { ConnectToCodeHostStep } from './steps/ConnectToCodeHostsStep'

import styles from './Setup.module.scss'

export const SetupWizard: FC = props => {
    const {} = props

    const [step, setStep] = useState(0)

    // Enforce the right class is added on the body for supporting different
    // themes based on user OS preferences
    const { enhancedThemePreference } = useTheme()
    const isLightTheme = enhancedThemePreference === ThemePreference.Light

    return (
        <div className={styles.root}>
            <BrandLogo variant="logo" isLightTheme={isLightTheme} className={styles.logo} />

            <Text className={styles.description}>Single docker setup - Version 4.4</Text>

            <SetupTabs activeTabIndex={step} defaultActiveIndex={0} onTabChange={setStep}>
                <SetupList>
                    <SetupTab index={0}>Connect to code hosts</SetupTab>
                    <SetupTab index={1}>Add Repositories</SetupTab>
                    <SetupTab index={2}>Finish</SetupTab>
                </SetupList>
                <SetupSteps>
                    <ConnectToCodeHostStep />
                    <AddRepositoriesStep />
                    <CloningRepositoriesStep />
                </SetupSteps>
            </SetupTabs>
        </div>
    )
}
