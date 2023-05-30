import { FC, useContext } from 'react'

import classNames from 'classnames'

import { Button, H1, H2, H3, Link, Text } from '@sourcegraph/wildcard'

import { tauriShellOpen } from '../../../../app/tauriIcpUtils'
import { SetupStepsContext, StepComponentProps } from '../../../../setup-wizard/components'

import styles from './AppInstallExtensionsSetupStep.module.scss'

export const AppInstallExtensionsSetupStep: FC<StepComponentProps> = ({ className }) => {
    const { onNextStep } = useContext(SetupStepsContext)

    return (
        <div className={classNames(styles.root, className)}>
            <div className={styles.content}>
                <div className={styles.description}>
                    <H1 className={styles.descriptionHeading}>Get Cody in your editor</H1>
                    <Text className={styles.descriptionText}>
                        Our extension combines an LLM with the context of your code to help you generate and fix code
                        more accurately.
                    </Text>
                </div>

                <div className={styles.actions}>
                    <div className={styles.actionsCard}>
                        <H3 as={H2}>You’ll need a Sourcegraph.com account in order to connect Cody.</H3>

                        <div className={styles.actionsButtonsGroup}>
                            <Button
                                size="lg"
                                variant="primary"
                                className={styles.actionsButton}
                                onClick={() => tauriShellOpen('vscode:extension/sourcegraph.cody-ai')}
                            >
                                VS Code Extension
                            </Button>

                            <Button
                                as={Link}
                                to="https://docs.sourcegraph.com/integration/editor"
                                target="_blank"
                                size="lg"
                                variant="secondary"
                            >
                                Other editors
                            </Button>

                            <Button variant="secondary" size="lg" className={styles.actionsButton} onClick={onNextStep}>
                                Next →
                            </Button>
                        </div>
                    </div>
                </div>
            </div>

            <img
                src="https://storage.googleapis.com/sourcegraph-assets/cody-extension.png"
                alt=""
                className={styles.image}
            />
        </div>
    )
}
