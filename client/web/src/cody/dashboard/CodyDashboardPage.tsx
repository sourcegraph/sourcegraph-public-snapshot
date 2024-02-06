import { type FC, useState } from 'react'

import { mdiChevronDown } from '@mdi/js'
import classNames from 'classnames'

import {
    Text,
    H1,
    ButtonLink,
    Link,
    Menu,
    MenuButton,
    MenuList,
    MenuItem,
    Icon,
    Position,
    MenuLink,
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
        name: 'VSCode',
        setupLink: 'https://docs.sourcegraph.com/cody/overview/install-vscode',
    },
    {
        icon: <IntelliJIcon className={styles.linkSelectorIcon} />,
        maker: 'Jetbrains',
        name: 'IntelliJ',
        setupLink: 'https://docs.sourcegraph.com/cody/overview/install-jetbrains',
    },
]

interface CodyDashboardPageProps {}

export const CodyDashboardPage: FC<CodyDashboardPageProps> = () => {
    const codySetupLink = 'https://docs.sourcegraph.com/cody/overview#getting-started'
    return (
        <section className={styles.dashboardContainer}>
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
                    <Text className={styles.dashboardText}>Download Cody for your favorite IDE</Text>
                    <LinkSelector options={setupOptions} />
                    <Text className="text-muted">
                        Struggling with setup?{' '}
                        <Link to={codySetupLink} className={styles.dashboardOnboardingIdeInstallationLink}>
                            Explore installation docs
                        </Link>
                        .
                    </Text>
                </section>
                <section className={styles.dashboardOnboardingWeb}>
                    <Text className={styles.dashboardText}>... or try it on the web</Text>
                    <ButtonLink to="/cody/chat" outline={true} className={styles.dashboardOnboardingWebLink}>
                        <CodyColorIcon className={styles.dashboardOnboardingCodyIcon} />
                        <span>Cody for web</span>
                    </ButtonLink>
                </section>
            </section>
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
                <MenuLink as={Link} className={styles.linkSelectorInfo} to={selectedOption.setupLink} target="_blank">
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
                    {options.map(option => (
                        <MenuItem className={styles.linkSelectorItem} onSelect={() => setSelectedOption(option)}>
                            <Text className="m-0">Install Cody on {option.name}</Text>
                        </MenuItem>
                    ))}
                </MenuList>
            </Menu>
        </section>
    )
}
