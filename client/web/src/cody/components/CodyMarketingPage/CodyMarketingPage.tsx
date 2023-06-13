import { mdiChevronRight, mdiCodeBracesBox, mdiGit, mdiMicrosoftVisualStudioCode } from '@mdi/js'
import classNames from 'classnames'

import { Theme, useTheme } from '@sourcegraph/shared/src/theme'
import { H1, H2, H3, H4, Icon, Link, PageHeader, Text } from '@sourcegraph/wildcard'

import { ExternalsAuth } from '../../../auth/components/ExternalsAuth'
import { MarketingBlock } from '../../../components/MarketingBlock'
import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { SourcegraphContext } from '../../../jscontext'
import { eventLogger } from '../../../tracking/eventLogger'
import { EventName } from '../../../util/constants'
import { CodyColorIcon, CodyHelpIcon, CodyWorkIcon } from '../../chat/CodyPageIcon'

import styles from './CodyMarketingPage.module.scss'

interface CodyPlatformCardProps {
    icon: string
    title: string
    description: string
    illustration: string
}

const onSpeakToAnEngineer = (): void => eventLogger.log(EventName.SPEAK_TO_AN_ENGINEER_CTA)
const onClickCTAButton = (type: string): void =>
    eventLogger.log('SignupInitiated', { type, source: 'cody-signed-out' }, { type, page: 'cody-signed-out' })

const codyPlatformCardItems = [
    {
        title: 'Knows your code',
        description:
            'Cody knows about your codebase and can use that knowledge to explain, generate, and improve your code.',
        icon: mdiCodeBracesBox,
        illustration: {
            dark: 'https://storage.googleapis.com/sourcegraph-assets/app-images/cody-knows-your-code-illustration-dark.png',
            light: 'https://storage.googleapis.com/sourcegraph-assets/app-images/cody-knows-your-code-illustration-light.png',
        },
    },
    {
        title: 'Works with VS Code',
        description:
            'This extension combines an LLM with the context of your code to help you generate and fix code more accurately.',
        icon: mdiMicrosoftVisualStudioCode,
        illustration: {
            dark: 'https://storage.googleapis.com/sourcegraph-assets/app-images/cody-vs-code-illustration-dark.png',
            light: 'https://storage.googleapis.com/sourcegraph-assets/app-images/cody-vs-code-illustration-light.png',
        },
    },
    {
        title: 'Try it on sourcegraph.com',
        description: 'Cody explains, generates, convert code, and more within the context of public repositories.',
        icon: mdiGit,
        illustration: {
            dark: 'https://storage.googleapis.com/sourcegraph-assets/app-images/cody-com-illustration-dark.png',
            light: 'https://storage.googleapis.com/sourcegraph-assets/app-images/cody-com-illustration-light.png',
        },
    },
]

export interface CodyMarketingPageProps {
    context: Pick<SourcegraphContext, 'authProviders'>
}

export const CodyMarketingPage: React.FunctionComponent<CodyMarketingPageProps> = ({ context }) => {
    const { theme } = useTheme()
    const isDarkTheme = theme === Theme.Dark

    return (
        <Page>
            <PageTitle title="Cody AI" />
            <PageHeader
                description={
                    <>
                        Cody answers code questions and writes code for you by reading your entire codebase and the code
                        graph.
                    </>
                }
                className={classNames('mb-3', styles.pageHeader)}
            >
                <PageHeader.Heading as="h2" styleAs="h1">
                    <PageHeader.Breadcrumb icon={CodyColorIcon}>
                        <div className={classNames('d-inline-flex align-items-center', styles.pageHeaderBreadcrumb)}>
                            Cody AI
                        </div>
                    </PageHeader.Breadcrumb>
                </PageHeader.Heading>
            </PageHeader>

            {/* Page content */}
            <div className={styles.headerSection}>
                <div>
                    <H1>Meet Cody, your AI assistant</H1>
                    <div className="ml-3">
                        <Text className={styles.codyConversation}>Auto-generate unit tests</Text>
                        <Text className={styles.codyConversation}>Explain code</Text>
                        <Text className={styles.codyConversation}>Find code smells</Text>
                        <Text className={styles.codyConversation}>Generate code</Text>
                    </div>
                    <CodyHelpIcon className={styles.codyHelpIcon} />
                </div>

                <MarketingBlock
                    contentClassName={styles.codySignUpPanel}
                    wrapperClassName={styles.codySignUpPanelWrapper}
                >
                    <H2>Sign up to get free access</H2>
                    <Text className="mt-3">
                        Cody combines an LLM with the context of Sourcegraph's code graph on public code or your code at
                        work. Sign up with:
                    </Text>
                    <div className={styles.buttonWrapper}>
                        <ExternalsAuth
                            context={context}
                            githubLabel="GitHub"
                            gitlabLabel="GitLab"
                            withCenteredText={true}
                            onClick={onClickCTAButton}
                            ctaClassName={styles.authButton}
                            iconClassName={styles.buttonIcon}
                        />
                    </div>
                    <Link to="https://sourcegraph.com/sign-up?showEmail=true">Or, continue with email</Link>
                    <Text className="mt-3 mb-0">
                        By registering, you agree to our{' '}
                        <Link
                            className={styles.termsPrivacyLink}
                            to="https://about.sourcegraph.com/terms"
                            target="_blank"
                            rel="noopener"
                        >
                            Terms of Service
                        </Link>{' '}
                        and{' '}
                        <Link
                            className={styles.termsPrivacyLink}
                            to="https://about.sourcegraph.com/privacy"
                            target="_blank"
                            rel="noopener"
                        >
                            Privacy Policy
                        </Link>
                        .
                    </Text>
                </MarketingBlock>
            </div>

            <div className={styles.learnMoreSection}>
                <H2>Enhancing productivity with Cody</H2>

                <div
                    className={classNames(
                        'd-flex flex-row flex-wrap mt-3 justify-content-center',
                        styles.startingPointWrapper
                    )}
                >
                    {codyPlatformCardItems.map(item => (
                        <CodyPlatformCard
                            key={item.title}
                            title={item.title}
                            description={item.description}
                            illustration={isDarkTheme ? item.illustration.dark : item.illustration.light}
                            icon={item.icon}
                        />
                    ))}
                </div>

                <div className={styles.learnMoreItemsWrapper}>
                    <div className={styles.learnMoreItems}>
                        <H4 className={styles.learnMoreItemsTitle}>Overview</H4>
                        <Text className="mb-0">
                            Visit the{' '}
                            <Link to="https://about.sourcegraph.com/cody" target="_blank" rel="noopener">
                                product page
                            </Link>{' '}
                            and see what devs are building with Cody.
                        </Text>
                    </div>

                    <div className={styles.learnMoreItems}>
                        <H4 className={styles.learnMoreItemsTitle}>Documentation</H4>
                        <Text className="mb-0">
                            Learn about Cody’s use cases, recipes, and FAQs on the{' '}
                            <Link to="/help/cody" target="_blank" rel="noopener">
                                documentation page
                            </Link>
                            .
                        </Text>
                    </div>
                </div>
            </div>

            <div className={styles.footer}>
                <CodyWorkIcon />
                <div>
                    <H1 className="mb-2">Get Cody for work</H1>
                    <Text className={styles.footerDescription}>
                        Cody for Sourcegraph Enterprise utilizes Sourcegraph's code graph to deliver context-aware
                        answers based on your private codebase, enabling enhanced code comprehension and productivity.
                    </Text>
                    <div className="mb-2">
                        <Link
                            to="https://about.sourcegraph.com/demo"
                            className={classNames('d-inline-flex align-items-center', styles.footerCtaLink)}
                            onClick={onSpeakToAnEngineer}
                        >
                            Speak to an engineer
                            <Icon svgPath={mdiChevronRight} aria-hidden={true} />
                        </Link>
                    </div>
                </div>
            </div>
        </Page>
    )
}

const CodyPlatformCard: React.FunctionComponent<CodyPlatformCardProps> = ({
    icon,
    title,
    description,
    illustration,
}) => (
    <div className={styles.codyPlatformCardWrapper}>
        <div className="d-flex flex-row align-items-center">
            <Icon svgPath={icon} aria-label="Close try Cody widget" className={styles.codyPlatformCardIcon} />
            <H3 className="ml-2 mb-0">{title}</H3>
        </div>
        <Text className={classNames('mt-2', styles.codyPlatformCardDescription)}>{description}</Text>
        <img src={illustration} alt={title} className={styles.codyPlatformCardImage} />
    </div>
)
