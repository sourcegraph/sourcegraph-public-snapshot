import { useEffect, useState, type FC } from 'react'

import { mdiChevronDown } from '@mdi/js'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import {
    ButtonLink,
    H1,
    H2,
    Icon,
    Link,
    Menu,
    MenuButton,
    MenuItem,
    MenuLink,
    MenuList,
    Position,
    Text,
} from '@sourcegraph/wildcard'

import { CodyColorIcon } from '../chat/CodyPageIcon'
import { IntelliJIcon } from '../upsell/IntelliJ'
import { VSCodeIcon } from '../upsell/vs-code'

import styles from './CodyDashboardPage.module.scss'

interface SetupOption {
    icon: JSX.Element
    maker: string
    name: string
    setupLink: string
}

const setupOptions: SetupOption[] = [
    {
        icon: <VSCodeIcon className={styles.linkSelectorIcon} />,
        maker: 'Microsoft',
        name: 'VS Code',
        setupLink: 'https://sourcegraph.com/docs/cody/clients/install-vscode',
    },
    {
        icon: <IntelliJIcon className={styles.linkSelectorIcon} />,
        maker: 'JetBrains',
        name: 'IntelliJ',
        setupLink: 'https://sourcegraph.com/docs/cody/clients/install-jetbrains',
    },
]

interface CodyDashboardPageProps extends TelemetryV2Props {}

export const CodyDashboardPage: FC<CodyDashboardPageProps> = ({ telemetryRecorder }) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('cody.dashboard', 'view')
    }, [telemetryRecorder])

    const codySetupLink = 'https://sourcegraph.com/docs/cody'
    return !window.context?.codyEnabledOnInstance ? (
        // This page should not be linked from anywhere if Cody is disabled on the instance, but add
        // a check here just in case to avoid confusing users if they find their way here.
        <section className={styles.dashboardContainer}>
            <section className={styles.dashboardHero}>
                <H1 className={styles.dashboardHeroHeader}>Cody is not enabled</H1>
                <Text className={styles.dashboardHeroTagline}>
                    Contact your Sourcegraph admin if this is unexpected.
                </Text>
            </section>
        </section>
    ) : (
        <section className={styles.dashboardContainer}>
            {window.context?.codyEnabledForCurrentUser ? (
                <>
                    <section className={styles.dashboardHero}>
                        <CodyColorIcon className={styles.dashboardCodyIcon} />
                        <H1 className={styles.dashboardHeroHeader}>
                            Get started with <span className={styles.codyGradient}>Cody</span>
                        </H1>
                        <Text className={styles.dashboardHeroTagline}>
                            Hey! 👋 Let’s get started with Cody — your new AI coding assistant.
                        </Text>
                    </section>
                    <section className={styles.dashboardOnboarding}>
                        <section className={styles.dashboardOnboardingIde}>
                            <Text className={styles.dashboardText}>Use Cody in your editor</Text>
                            <LinkSelector options={setupOptions} />
                            <Text className="text-muted">
                                <Link to={codySetupLink} className={styles.dashboardOnboardingIdeInstallationLink}>
                                    Documentation
                                </Link>
                            </Text>
                        </section>
                        <section className={styles.dashboardOnboardingWeb}>
                            <Text className={styles.dashboardText}>... or try it on the web</Text>
                            <ButtonLink to="/cody/chat" outline={true} className={styles.dashboardOnboardingWebLink}>
                                <CodyColorIcon className={styles.dashboardOnboardingCodyIcon} />
                                <span>Cody Web</span>
                            </ButtonLink>
                        </section>
                    </section>
                </>
            ) : (
                <section className={styles.dashboardHero}>
                    <CodyColorIcon className={styles.dashboardCodyIcon} />
                    <H2 className={styles.dashboardHeroHeader}>
                        Your user account doesn't have access to <span className={styles.codyGradient}>Cody</span>
                    </H2>
                    <Text className={styles.dashboardHeroTagline}>
                        Ask your Sourcegraph admin to{' '}
                        <Link to="/help/cody/clients/enable-cody-enterprise#enable-cody-only-for-some-users">
                            enable Cody for you
                        </Link>
                    </Text>
                </section>
            )}
        </section>
    )
}

interface LinkSelectorProps {
    options: SetupOption[]
}

const LinkSelector: FC<LinkSelectorProps> = ({ options }) => {
    const [firstOption] = options
    const [selectedOption, setSelectedOption] = useState<SetupOption>(firstOption)
    return (
        <section className={styles.linkSelectorContainer}>
            <Menu>
                <MenuLink
                    as={Link}
                    className={styles.linkSelectorInfo}
                    to={selectedOption.setupLink}
                    target="_blank"
                    rel="noreferrer"
                >
                    {selectedOption.icon}
                    <section>
                        <Text className={styles.linkSelectorOptionMaker}>{selectedOption.maker}</Text>
                        <Text className={styles.linkSelectorOptionName}>{selectedOption.name}</Text>
                    </section>
                </MenuLink>
                <MenuButton variant={undefined} className={styles.linkSelectorBtn}>
                    <Icon size="md" aria-hidden={true} svgPath={mdiChevronDown} />
                </MenuButton>

                <MenuList position={Position.bottomEnd} className={styles.linkSelectorDropdown}>
                    {options.map((option, index) => (
                        <MenuItem
                            key={index}
                            className={styles.linkSelectorItem}
                            onSelect={() => setSelectedOption(option)}
                        >
                            <Text className="m-0">Install Cody on {option.name}</Text>
                        </MenuItem>
                    ))}
                </MenuList>
            </Menu>
        </section>
    )
}
