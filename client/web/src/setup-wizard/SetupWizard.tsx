import { FC, useState } from 'react'

import { H1, H2 } from '@sourcegraph/wildcard'

import { BrandLogo } from '../components/branding/BrandLogo'
import { SiteAdminRepositoriesContainer } from '../site-admin/SiteAdminRepositoriesPage'

import { SetupSteps, SetupTabs, SetupList, SetupTab } from './components/SetupTabs'
import { Text } from '@sourcegraph/wildcard'

import styles from './Setup.module.scss'

export const SetupWizard: FC = props => {
    const [step, setStep] = useState(0)

    return (
        <div className={styles.root}>
            <header className={styles.header}>
                <BrandLogo variant="logo" isLightTheme={false} className={styles.logo} />

                <H2 as={H1} className="font-weight-normal text-white mt-3 mb-4">
                    Welcome to Sourcegraph! Let's get your instance ready.
                </H2>
            </header>

            <SetupTabs activeTabIndex={step} defaultActiveIndex={0} onTabChange={setStep}>
                <SetupList wrapperClassName="border-bottom-0">
                    <SetupTab index={0}>Add code hosts</SetupTab>
                    <SetupTab index={1}>Sync repositories</SetupTab>
                </SetupList>
            </SetupTabs>
            <SetupSteps>
                <SiteAdminRepositoriesContainer>
                    <Text>
                        It may take a few moments to clone and index each repository. Repository statuses are displayed
                        below.
                    </Text>
                </SiteAdminRepositoriesContainer>
            </SetupSteps>
        </div>
    )
}
