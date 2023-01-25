import { FC, useState } from 'react'

import { H1, H2 } from '@sourcegraph/wildcard'

import { BrandLogo } from '../components/branding/BrandLogo'

import { SetupTabs, SetupList, SetupTab } from './components/SetupTabs'

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
                    <SetupTab index={0}>Connect your code</SetupTab>
                    <SetupTab index={1}>Add Repositories</SetupTab>
                    <SetupTab index={2}>Start searching</SetupTab>
                </SetupList>
            </SetupTabs>
        </div>
    )
}
