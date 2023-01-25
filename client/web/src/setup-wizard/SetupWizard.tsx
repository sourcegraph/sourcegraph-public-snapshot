import { FC, useState } from 'react'
import { useNavigate } from 'react-router-dom-v5-compat'

import { H1, H2 } from '@sourcegraph/wildcard'

import { BrandLogo } from '../components/branding/BrandLogo'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import { ThemePreference, useTheme } from '../theme'

import { SetupTabs, SetupList, SetupTab } from './components/SetupTabs'

import styles from './Setup.module.scss'

export const SetupWizard: FC = props => {
    const {} = props

    const [isSetupWizardEnabled] = useFeatureFlag('local-setup-wizard')
    const navigate = useNavigate()
    const [step, setStep] = useState(0)

    if (!isSetupWizardEnabled) {
        navigate('/')
    }

    // Enforce the right class is added on the body for supporting different
    // themes based on user OS preferences
    const { enhancedThemePreference } = useTheme()
    const isLightTheme = enhancedThemePreference === ThemePreference.Light

    return (
        <div className={styles.root}>
            <BrandLogo variant="logo" isLightTheme={isLightTheme} className={styles.logo} />

            <H2 as={H1} className={styles.description}>
                Welcome to Sourcegraph! Let's get your instance ready.
            </H2>

            <SetupTabs activeTabIndex={step} defaultActiveIndex={0} onTabChange={setStep}>
                <SetupList>
                    <SetupTab index={0}>Connect to code hosts</SetupTab>
                    <SetupTab index={1}>Add Repositories</SetupTab>
                    <SetupTab index={2}>Finish</SetupTab>
                </SetupList>
            </SetupTabs>
        </div>
    )
}
