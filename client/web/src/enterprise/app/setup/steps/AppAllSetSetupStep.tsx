import { type FC, useCallback, useContext, useState } from 'react'

import { readTextFile, BaseDirectory } from '@tauri-apps/api/fs'
import classNames from 'classnames'
import { useSearchParams } from 'react-router-dom'

import { Button, H1, Text, Tooltip } from '@sourcegraph/wildcard'

import { FooterWidget, SetupStepsContext, type StepComponentProps } from '../../../../setup-wizard/components'
import { AppSetupProgressBar } from '../components/AppSetupProgressBar'

import styles from './AppAllSetSetupStep.module.scss'

export const AppAllSetSetupStep: FC<StepComponentProps> = ({ className }) => {
    const { onNextStep } = useContext(SetupStepsContext)
    const [isProgressFinished, setProgressFinished] = useState(false)
    const [searchParams] = useSearchParams()

    const handleOneRepositoryFinished = useCallback(() => {
        localStorage.setItem('app.setup.finished', 'true')
        setProgressFinished(true)
    }, [])

    const fromVSCode = searchParams.get('from') === 'vscode'

    return (
        <div className={classNames(styles.root, className)}>
            <div className={styles.description}>
                <H1 className={styles.descriptionHeading}>
                    {isProgressFinished ? 'Embeddings generation complete 🎉' : 'Generating embeddings…'}
                </H1>

                <div className={styles.descriptionContent}>
                    <Text className={styles.descriptionText}>
                        {isProgressFinished
                            ? 'If you need to create embeddings for additional repositories, open Cody App and use Settings → Add a repository.'
                            : 'Once embeddings generation is complete you can return to VS Code and Cody will answer questions with additional context of your entire codebase.'}
                    </Text>

                    <Tooltip content={!isProgressFinished ? 'The code graph is still being generated' : ''}>
                        <Button
                            size="lg"
                            variant="primary"
                            disabled={!isProgressFinished}
                            className={styles.descriptionButton}
                            onClick={async () => {
                                if (!fromVSCode) {
                                    onNextStep()
                                    return
                                }

                                let redirect = 'vscode://sourcegraph.cody-ai/app-done'

                                try {
                                    const config = await readTextFile('vscode.json', { dir: BaseDirectory.AppData })

                                    redirect = JSON.parse(config).redirect || redirect
                                } catch {
                                    // noop
                                }

                                ;(window as any).__TAURI__.shell.open(redirect)
                                onNextStep()
                            }}
                        >
                            {fromVSCode ? 'Return to VS Code' : 'Get Started'}
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
