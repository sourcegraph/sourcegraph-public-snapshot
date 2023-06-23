import { FC, useCallback, useContext, useState } from 'react'

import classNames from 'classnames'

import { Button, H1, Text, Tooltip } from '@sourcegraph/wildcard'

import { FooterWidget, SetupStepsContext, StepComponentProps } from '../../../../setup-wizard/components'
import { AppSetupProgressBar } from '../components/AppSetupProgressBar'

import styles from './AppAllSetSetupStep.module.scss'

export const AppAllSetSetupStep: FC<StepComponentProps> = ({ className }) => {
    const { onNextStep } = useContext(SetupStepsContext)
    const [isProgressFinished, setProgressFinished] = useState(false)

    const handleOneRepositoryFinished = useCallback(() => {
        setProgressFinished(true)
    }, [])

    return (
        <div className={classNames(styles.root, className)}>
            <div className={styles.description}>
                <H1 className={styles.descriptionHeading}>You’re all set</H1>

                <div className={styles.descriptionContent}>
                    <Text className={styles.descriptionText}>
                        Once embeddings are finished being generated, you can specify Cody’s context and start asking
                        cody questions in Cody Chat.
                    </Text>

                    <Tooltip content={!isProgressFinished ? 'Embeddings are still being generated' : ''}>
                        <Button
                            size="lg"
                            variant="primary"
                            disabled={!isProgressFinished}
                            className={styles.descriptionButton}
                            onClick={onNextStep}
                        >
                            Get Started
                        </Button>
                    </Tooltip>
                </div>
            </div>

            <div className={styles.imageWrapper}>
                <img
                    src="https://storage.googleapis.com/sourcegraph-assets/setup/vscode-and-cody-chat.png"
                    alt=""
                    className={styles.image}
                />
            </div>

            <FooterWidget>
                <AppSetupProgressBar onOneRepositoryFinished={handleOneRepositoryFinished} className={styles.footer} />
            </FooterWidget>
        </div>
    )
}
