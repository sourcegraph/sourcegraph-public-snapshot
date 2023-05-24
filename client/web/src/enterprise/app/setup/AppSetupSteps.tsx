import { FC, useContext } from 'react'

import { mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { H1, H2, H3, Text, Link, Button, Icon } from '@sourcegraph/wildcard'

import { tauriShellOpen } from '../../../app/tauriIcpUtils'
import { EnterprisePageRoutes } from '../../../routes.constants'
import { SetupStepsContext } from '../../../setup-wizard/components'
import { LocalRepositoriesWidget, PathsPickerActions } from '../settings/local-repositories/LocalRepositoriesTab'

import styles from './AppSetupSteps.module.scss'

export const AppWelcomeSetupPage: FC = () => {
    const { onNextStep } = useContext(SetupStepsContext)

    return (
        <div className={styles.root}>
            <div className={styles.description}>
                <H1 className={styles.descriptionHeading}>Welcome</H1>
                <Text className={styles.descriptionText}>
                    The Sourcegraph app now includes our AI coding assistant Cody, which helps you read, write, and
                    understand code 10x faster.
                </Text>

                <Text className={styles.descriptionFooter}>
                    <Link to="https://docs.sourcegraph.com/cody/quickstart" rel="noopener" target="_blank">
                        Learn more
                    </Link>{' '}
                    about how Cody works.
                </Text>

                <img
                    src="https://storage.googleapis.com/sourcegraph-assets/cody-picture-full.png"
                    alt=""
                    className={styles.descriptionImage}
                />
            </div>

            <div className={styles.actions}>
                <div className={styles.actionsCard}>
                    <H3 as={H2}>You’ll need a Sourcegraph.com account in order to connect Cody.</H3>

                    <div className={styles.actionsButtons}>
                        <Button
                            as={Link}
                            to={`https://sourcegraph.com/user/settings/tokens/new/callback?requestFrom=APP&destination=${EnterprisePageRoutes.AppSetup}/local-repositories`}
                            variant="primary"
                            size="lg"
                            className={styles.actionsButton}
                            target="_blank"
                        >
                            <SourcegraphLogo />
                            Connect to Sourcegraph.com
                        </Button>

                        <Button variant="secondary" size="lg" className={styles.actionsButton} onClick={onNextStep}>
                            Skip →
                        </Button>
                    </div>

                    <Text className={styles.actionsFooter}>
                        By signing in, you’re connecting Cody to your app and agreeing to{' '}
                        <Link
                            to="https://about.sourcegraph.com/terms/cody-notice"
                            target="_blank"
                            rel="noopener"
                            className={styles.actionsTermLink}
                        >
                            Sourcegraph’s Cody Usage Privacy Notice
                        </Link>
                        .
                    </Text>
                </div>
            </div>
        </div>
    )
}

const SourcegraphLogo: FC = () => (
    <svg width="21" height="20" viewBox="0 0 17 16" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path
            d="M12.1978 2.55945C12.4181 2.37078 12.6756 2.25634 12.9394 2.20685C13.4793 2.10788 14.0596 2.29036 14.4413 2.73885C15.0122 3.40693 14.9315 4.40906 14.2644 4.97818L11.9061 6.98244L10.1746 6.37002L9.52917 6.14114L8.44312 5.75761L9.32127 5.0122L9.84258 4.5699L12.2009 2.56564L12.1978 2.55945Z"
            fill="white"
        />
        <path
            d="M4.30672 13.4406C4.08641 13.6292 3.82886 13.7437 3.5651 13.7932C3.02518 13.8921 2.44491 13.7097 2.06324 13.2612C1.49228 12.5931 1.57296 11.5909 2.24011 11.0218L4.59841 9.01758L6.32989 9.62999L6.97532 9.85887L8.06138 10.2424L7.18323 10.9878L6.66192 11.4301L4.30362 13.4344L4.30672 13.4406Z"
            fill="white"
        />
        <path
            d="M5.49827 1.87587C5.34002 1.01293 5.91097 0.187097 6.77361 0.0262615C7.63936 -0.131481 8.46786 0.43763 8.62922 1.29748L9.19087 4.3348L7.7914 5.5256L6.05992 4.91319L5.49827 1.87587Z"
            fill="white"
        />
        <path
            d="M11.0061 14.1241C11.1643 14.987 10.5934 15.8129 9.73073 15.9737C8.86499 16.1315 8.03648 15.5623 7.87512 14.7025L7.31348 11.6652L8.71294 10.4744L10.4444 11.0868L11.0061 14.1241Z"
            fill="white"
        />
        <path
            d="M15.9182 10.7095C15.7196 11.2632 15.2479 11.6405 14.7049 11.7395C14.438 11.789 14.1588 11.7704 13.8857 11.6745L10.9658 10.6415L10.3203 10.4126L9.23427 10.029L8.58884 9.80017L6.85736 9.18775L6.21193 8.95887L5.12587 8.57534L4.48044 8.34646L1.5605 7.3134C0.731993 7.01956 0.29757 6.11332 0.592357 5.28749C0.790951 4.73384 1.26261 4.3565 1.80564 4.25752C2.0725 4.20803 2.35177 4.22659 2.62484 4.32247L5.54478 5.35553L6.19021 5.58441L7.27627 5.96795L7.92169 6.19683L9.65318 6.80924L10.2986 7.03812L11.3847 7.42165L12.0301 7.65053L14.95 8.68359C15.7785 8.97743 16.213 9.88368 15.9182 10.7095V10.7095Z"
            fill="white"
        />
    </svg>
)

export const AddLocalRepositoriesSetupPage: FC = () => {
    const { onPrevStep, onNextStep } = useContext(SetupStepsContext)

    return (
        <div className={styles.root}>
            <div className={styles.description}>
                <H1 className={styles.descriptionHeading}>Add your projects</H1>

                <Text className={styles.descriptionText}>
                    Choose the local repositories you’d like to add to your app.
                </Text>

                <Text className={styles.descriptionText}>
                    The app will automatically create embeddings to give Cody better context on your code.
                </Text>

                <Text className={styles.descriptionText}>
                    You can sync, unsync, or connect remote repos later, under Settings.
                </Text>

                <footer className={styles.localRepositoriesNavigation}>
                    <Button variant="primary" size="lg" onClick={onPrevStep}>
                        ←
                    </Button>
                    <Button variant="primary" size="lg" onClick={onNextStep}>
                        Next →
                    </Button>
                </footer>
            </div>
            <div className={classNames(styles.actions, styles.actionsWithBg)}>
                <LocalRepositoriesWidget className={styles.localRepositoriesWidget}>
                    {api => (
                        <PathsPickerActions
                            className={styles.localRepositoriesButtonsGroup}
                            onPathsChange={api.addNewPaths}
                        />
                    )}
                </LocalRepositoriesWidget>
            </div>
        </div>
    )
}

export const InstallExtensionsSetupPage: FC = () => {
    const { onPrevStep, onNextStep } = useContext(SetupStepsContext)

    return (
        <div className={styles.root}>
            <div className={styles.extension}>
                <H1 className={styles.extensionHeading}>Get Cody in your editor</H1>

                <Text className={styles.descriptionText}>Get Cody in your editor by installing Cody extension.</Text>

                <footer className={styles.extensionFooter}>
                    <Button variant="primary" size="lg" onClick={onPrevStep} className={styles.actionsButton}>
                        ←
                    </Button>

                    <Button
                        variant="primary"
                        size="lg"
                        className={styles.actionsButton}
                        onClick={() => tauriShellOpen('vscode:extension/sourcegraph.cody-ai')}
                    >
                        <Icon svgPath={mdiOpenInNew} aria-hidden={true} />
                        Install VSCode extension
                    </Button>

                    <Button
                        as={Link}
                        to="https://docs.sourcegraph.com/integration/editor"
                        variant="primary"
                        size="lg"
                        target="_blank"
                        className={styles.actionsButton}
                    >
                        <Icon svgPath={mdiOpenInNew} aria-hidden={true} />
                        Other editors
                    </Button>

                    <Button variant="secondary" size="lg" onClick={onNextStep} className={styles.actionsButton}>
                        Skip, I'll do it later →
                    </Button>
                </footer>
            </div>
        </div>
    )
}
