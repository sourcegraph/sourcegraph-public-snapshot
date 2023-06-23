import { useEffect, useRef } from 'react'

import { mdiEmailOutline, mdiMicrosoftVisualStudioCode, mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { Badge, H2, Icon, Link, PageHeader, Text } from '@sourcegraph/wildcard'

import { ExternalsAuth } from '../auth/components/ExternalsAuth'
import { CodyLetsWorkIcon, CodyStartCoding } from '../cody/chat/CodyPageIcon'
import { Page } from '../components/Page'
import { PageTitle } from '../components/PageTitle'
import { SourcegraphContext } from '../jscontext'
import { eventLogger } from '../tracking/eventLogger'
import { EventName } from '../util/constants'

import { BackgroundImage, Light, IntellijIcon, EmacsIcon, NeovimIcon } from './components/GetCodyPageIcon'

import styles from './GetCodyPage.module.scss'

interface WaitListButtonProps {
    title: string
    to: string
    icon: React.ReactElement
}

interface GetCodyPageProps {
    authenticatedUser: AuthenticatedUser | null
    context: Pick<SourcegraphContext, 'authProviders'>
}

const SOURCEGRAPH_MAC_DMG =
    'https://github.com/sourcegraph/sourcegraph/releases/download/app-v2023.6.13%2B1311.1af08ae774/Sourcegraph_2023.6.13+1311.1af08ae774_aarch64.dmg'

const onClickCTAButton = (type: string): void =>
    eventLogger.log(EventName.TRY_CODY_SIGNUP_INITIATED, { type, source: 'get-cody' }, { type, page: 'get-cody' })

export const GetCodyPage: React.FunctionComponent<GetCodyPageProps> = ({ authenticatedUser, context }) => {
    const lightRef = useRef<HTMLDivElement>(null)
    const scrollContainer = document.querySelector('main')

    useEffect(() => {
        if (!scrollContainer || !lightRef.current) {
            return
        }

        const handleScroll = (): void => {
            const lightElement = lightRef.current
            const alignedElements = document.querySelectorAll('.get-cody-step')

            if (lightElement) {
                const lightBounds = lightElement.getBoundingClientRect()

                for (const alignedElement of alignedElements) {
                    const elementBounds = alignedElement.getBoundingClientRect()

                    if (elementBounds.top <= lightBounds.top && elementBounds.bottom >= lightBounds.top) {
                        alignedElement.classList.add(styles.focusBackground)
                    } else {
                        alignedElement.classList.remove(styles.focusBackground)
                    }
                }
            }
        }

        scrollContainer.addEventListener('scroll', handleScroll)

        return () => {
            scrollContainer.removeEventListener('scroll', handleScroll)
        }
    }, [scrollContainer])

    return (
        <div className={styles.pageWrapper}>
            <Page className={styles.page}>
                <PageTitle title="Get Started with Cody" />
                <PageHeader className={styles.pageHeader}>
                    <Link to="https://about.sourcegraph.com/">
                        <img
                            className={styles.pageHeaderImage}
                            src="https://sourcegraph.com/.assets/img/sourcegraph-logo-dark.svg"
                            alt="Sourcegraph logo"
                        />
                    </Link>
                </PageHeader>

                <div className={styles.getStartedWithCody}>
                    <div>
                        <Text className={styles.getStartedWithCodyTitle}>Get started with Cody</Text>
                        <Text className={styles.getStartedWithCodyDescription}>
                            Try Cody free on your local machine with the Cody app and IDE extensions.
                        </Text>
                    </div>
                    <CodyLetsWorkIcon className={styles.codyLetsWorkImage} />
                </div>

                <div className={styles.cardWrapper}>
                    <div className={styles.dashedLine} />
                    <div className={styles.lightLine} />
                    <div ref={lightRef} className={styles.lightWrapper}>
                        <Light className={styles.light} />
                    </div>

                    {/* connect to cody section */}
                    {authenticatedUser && (
                        <div className={classNames(styles.card, 'get-cody-step', styles.focusBackground)}>
                            <H2 className={classNames(styles.cardTitle, 'mb-4')}>
                                You’ll need a Sourcegraph account to connect to Cody
                            </H2>
                            <Text className={classNames(styles.cardDescription, 'mb-4')}>
                                Log in or Sign up for a Sourcegraph.com account.
                            </Text>
                            <div className={styles.authButtonsWrap}>
                                <div className={styles.externalAuthWrapper}>
                                    <ExternalsAuth
                                        context={context}
                                        githubLabel="Github"
                                        gitlabLabel="Gitlab"
                                        withCenteredText={true}
                                        onClick={onClickCTAButton}
                                        ctaClassName={styles.authButton}
                                        iconClassName={styles.buttonIcon}
                                    />
                                </div>
                                <Link
                                    to="https://sourcegraph.com/sign-up?showEmail=true&returnTo=get-cody"
                                    className={classNames('text-decoration-none', styles.emailAuthButton)}
                                    onClick={() => onClickCTAButton('builtin')}
                                >
                                    <Icon
                                        className="mr-1"
                                        svgPath={mdiEmailOutline}
                                        inline={false}
                                        aria-hidden={true}
                                    />
                                    Continue with email
                                </Link>
                            </div>
                            <Text className={styles.terms}>
                                By registering, you agree to our{' '}
                                <Link
                                    to="https://about.sourcegraph.com/terms"
                                    className={styles.termsLink}
                                    target="_blank"
                                    rel="noopener"
                                >
                                    Terms of Service
                                </Link>{' '}
                                and{' '}
                                <Link
                                    to="https://about.sourcegraph.com/terms/privacy"
                                    className={styles.termsLink}
                                    target="_blank"
                                    rel="noopener"
                                >
                                    Privacy Policy
                                </Link>
                            </Text>
                        </div>
                    )}

                    {/* Install cody desktop app section */}
                    <div className={classNames(styles.card, 'get-cody-step')}>
                        <H2 className={classNames(styles.cardTitle, 'mb-4')}>
                            Install the <span className={styles.installCodyTitle}>Cody desktop app</span>
                        </H2>
                        <Text className={styles.cardDescription}>
                            The Cody app, when combined with a Cody IDE extension, enables context fetching for all of
                            your local repositories. Without the app, Cody only fetches context on the repository
                            currently open in your IDE.
                        </Text>
                        <Link
                            to={SOURCEGRAPH_MAC_DMG}
                            className={classNames('text-decoration-none', styles.downloadForMacButton)}
                        >
                            Download for Mac <Badge className={classNames('ml-2', styles.betaBadge)}>Beta</Badge>
                        </Link>
                        <Text className={styles.appleSilicon}>
                            Apple Silicon required.{' '}
                            <Link
                                to="https://docs.sourcegraph.com/app"
                                className={styles.appleSiliconLink}
                                target="_blank"
                                rel="noopener"
                            >
                                Other options.
                            </Link>
                        </Text>
                        <Text className={styles.terms}>
                            By using Sourcegraph, you agree to the{' '}
                            <Link
                                to="https://about.sourcegraph.com/terms/privacy"
                                className={styles.termsLink}
                                target="_blank"
                                rel="noopener"
                            >
                                privacy policy
                            </Link>{' '}
                            and{' '}
                            <Link
                                to="https://about.sourcegraph.com/terms"
                                className={styles.termsLink}
                                target="_blank"
                                rel="noopener"
                            >
                                terms
                            </Link>
                        </Text>
                    </div>

                    {/* Install cody extension section */}
                    <div className={classNames(styles.card, 'get-cody-step')}>
                        <H2 className={classNames(styles.cardTitle, 'mb-4')}>
                            Install the Cody extension for your IDE(s)
                        </H2>
                        <Text className={styles.cardDescription}>
                            If you’ve downloaded the app, it will prompt you to sign in to your Sourcegraph.com account,
                            connect your repositories, and connect your IDE extensions.
                        </Text>
                        <Link
                            to="vscode:extension/sourcegraph.cody-ai"
                            className={classNames('text-decoration-none', styles.downloadForVscode)}
                        >
                            <Icon
                                className={styles.vscodeIcon}
                                svgPath={mdiMicrosoftVisualStudioCode}
                                inline={false}
                                aria-hidden={true}
                            />{' '}
                            <span className={styles.downloadForVscodeText}>Install Cody for VS Code</span>
                        </Link>
                        <Link
                            to="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai#:~:text=Cody%20for%20VS%20Code%20is,not%20just%20your%20open%20files"
                            className={classNames('text-decoration-none', styles.vscodeMarketplace)}
                        >
                            Or download on the VS Code Marketplace
                            <Icon className="ml-2" svgPath={mdiChevronRight} inline={false} aria-hidden={true} />
                        </Link>
                        <div className={styles.comingSoonWrapper}>
                            <Text className={styles.comingSoonWrapperText}>Coming soon:</Text>
                            <div className={styles.joinWaitlistButtonWrapper}>
                                <WaitListButton
                                    to=""
                                    icon={<IntellijIcon className={styles.joinWaitlistButtonIcon} />}
                                    title="IntelliJ"
                                />
                                <WaitListButton
                                    to=""
                                    icon={<NeovimIcon className={styles.joinWaitlistButtonIcon} />}
                                    title="Neovim"
                                />
                                <WaitListButton
                                    to=""
                                    icon={<EmacsIcon className={styles.joinWaitlistButtonIcon} />}
                                    title="Emacs"
                                />
                            </div>
                        </div>
                    </div>

                    <div className={classNames(styles.card, styles.startCodingWithMeImgWrapper, 'get-cody-step')}>
                        <CodyStartCoding className={styles.startCodingWithMeImg} />
                    </div>

                    {/* cody for enterprise section */}
                    <div className={styles.codyForEnterprise}>
                        <H2 className={classNames(styles.cardTitle, styles.title)}>
                            Interested in Cody for Sourcegraph Enterprise?
                        </H2>
                        <Text className={styles.cardDescription}>
                            If you’ve downloaded the app, it will prompt you to sign in to your Sourcegraph.com account,
                            connect your repositories, and connect your IDE extensions.
                        </Text>
                        <Text className={classNames(styles.cardDescription, styles.getInTouchText)}>
                            Get in touch with our team to try Cody for Sourcegraph Enterprise.
                        </Text>
                        <div className={styles.ctaButtonWrapper}>
                            <Link
                                to="https://about.sourcegraph.com/demo"
                                className={classNames('text-decoration-none', styles.getCodyEnterpriseButton)}
                            >
                                Get Cody for Enterprise
                            </Link>
                            <Link
                                to="https://info.sourcegraph.com/talk-to-a-developer"
                                className={classNames('text-decoration-none', styles.talkToEngineerButton)}
                            >
                                Talk to an engineer
                            </Link>
                        </div>
                    </div>
                </div>

                <BackgroundImage className={styles.backgroundImage} />

                {/* footer section */}
                <div className={styles.footer}>
                    <Text>© 2023 Sourcegraph, Inc. </Text>
                    <Link
                        to="https://about.sourcegraph.com/terms"
                        className={styles.footerLink}
                        target="_blank"
                        rel="noopener"
                    >
                        Terms
                    </Link>
                    <Link
                        to="https://about.sourcegraph.com/security"
                        className={styles.footerLink}
                        target="_blank"
                        rel="noopener"
                    >
                        Security
                    </Link>
                    <Link
                        to="https://about.sourcegraph.com/terms/privacy"
                        className={styles.footerLink}
                        target="_blank"
                        rel="noopener"
                    >
                        Privacy
                    </Link>
                </div>
            </Page>
        </div>
    )
}

const WaitListButton: React.FunctionComponent<WaitListButtonProps> = ({ title, to, icon }) => (
    <Link to={to} className={classNames('text-decoration-none', styles.joinWaitlistButton)}>
        {icon}
        <span className={styles.joinWaitlistButtonTitle}>{title}</span>
        <span className={styles.joinWaitlistButtonLinkText}>
            Join the waitlist
            <Icon className="ml-2" svgPath={mdiChevronRight} inline={false} aria-hidden={true} />
        </span>
    </Link>
)
