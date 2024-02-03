import type { FC } from 'react'

import { mdiOpenInNew } from '@mdi/js'

import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { H2, Text, ButtonLink, Link, Icon } from '@sourcegraph/wildcard'

import { CodeSearchIcon } from './CodeSearchIcon'
import { FeatureImage } from './Feature'
import { IntegrationsIcon } from './IntegrationsIcon'
import { SearchExample } from './SearchExample'

import styles from './SearchUpsellPage.module.scss'

interface SearchFeature {
    title: string
    description: string
}

const searchFeatures: SearchFeature[] = [
    {
        title: 'Reuse high-quality code',
        description: 'Find code across thousands of repositories and multiple code hosts in seconds.',
    },
    {
        title: 'Resolve issues and incidents faster',
        description: 'Pinpoint root causes with symbol, commit, and diff searches.',
    },
    {
        title: 'Exhaustive search',
        description:
            "Discover every instance of vulnerable or buggy code in milliseconds and have complete confidence in what's in your codebase.",
    },
]

export const SearchUpsellPage: FC = () => {
    const isLightTheme = useIsLightTheme()
    const contactSalesLink = 'https://sourcegraph.com/contact/request-info'
    const findOutMoreLink = 'https://sourcegraph.com/code-search'
    return (
        <div className={styles.container}>
            <section className={styles.hero}>
                <CodeSearchIcon isLightTheme={isLightTheme} />
                <H2 className={styles.heroHeader}>Grok your entire codebase</H2>
                <Text className={styles.heroDescription}>
                    Code Search, along with complementary tools, helps devs find, fix, and onboard to new code quickly.
                </Text>

                <div>
                    <ButtonLink href={contactSalesLink} variant="primary" className="py-2 px-3 rounded mr-5">
                        Talk to a product expert
                    </ButtonLink>

                    <ButtonLink href={findOutMoreLink} variant="secondary" className="py-2 px-3 rounded">
                        Find out more
                    </ButtonLink>
                </div>
            </section>
            <SearchExample isLightTheme={isLightTheme} className={styles.searchExample} />

            <section className={styles.features}>
                <section className={styles.featuresMeta}>
                    <div>
                        <CodeSearchIcon isLightTheme={isLightTheme} className={styles.featuresCodeSearchIcon} />
                        <Text className={styles.featuresTagLine}>
                            Find and fix code in any code host, language, or repository
                        </Text>
                    </div>
                    <FeatureImage className={styles.featuresImage} />
                </section>
                <section className={styles.featuresGrid}>
                    {searchFeatures.map(({ title, description }, index) => (
                        <div key={index} className={styles.featuresCard}>
                            <Text className={styles.featuresCardTitle}>{title}</Text>
                            <Text className={styles.featuresCardDescription}>{description}</Text>
                        </div>
                    ))}
                </section>
            </section>

            <section className={styles.integrations}>
                <section className={styles.integrationsMeta}>
                    <Text className={styles.integrationsHeader}>Code Search integrates with Cody 🤝</Text>
                    <Text className={styles.integrationsDescription}>
                        Use Cody in Code Search to explain code, generate unit tests, transpile code, improve variable
                        names and a ton more!
                    </Text>
                </section>

                <IntegrationsIcon />
            </section>

            <section className={styles.otherIntegrations}>
                <div className={styles.otherIntegrationsGrid}>
                    <Text className={styles.otherIntegrationsTitle}>Understand your code and its dependencies</Text>
                    <Text className={styles.otherIntegrationsDescription}>
                        Complete code reviews, get up to speed on unfamiliar code, and determine the impact of code
                        changes with the confidence of compiler-accurate code navigation.
                    </Text>
                    <Link to="/help/code_navigation/explanations/introduction_to_code_navigation">
                        Find out more about Code Navigation →
                    </Link>
                </div>

                <section className={styles.otherIntegrationsGrid}>
                    <Text className={styles.otherIntegrationsTitle}>Automate large-scale code changes</Text>
                    <Text className={styles.otherIntegrationsDescription}>
                        Find all occurrences of code to change with Code Search and programmatically make those changes
                        by creating a declarative specification file.
                    </Text>
                    <Link to="https://sourcegraph.com/case-studies/indeed-accelerates-development-velocity">
                        Read how Indeed uses Batch Changes to accelerate deployment →
                    </Link>
                </section>
            </section>

            <section className={styles.footer}>
                <Text className={styles.footerText}>Code Search also works great with...</Text>
                <Link to="/help/code_monitoring">
                    Code Monitoring{' '}
                    <Icon svgPath={mdiOpenInNew} inline={false} aria-label="Learn more about Code Monitoring" />
                </Link>
                <Link to="/help/code_insights">
                    Insights <Icon svgPath={mdiOpenInNew} inline={false} aria-label="Learn more about Code Insights" />
                </Link>
                <Link to="/help/notebooks">
                    Notebooks <Icon svgPath={mdiOpenInNew} inline={false} aria-label="Learn more about Notebooks" />
                </Link>
            </section>
        </div>
    )
}
