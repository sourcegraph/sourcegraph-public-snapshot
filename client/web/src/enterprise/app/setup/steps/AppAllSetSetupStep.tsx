import { FC, useContext } from 'react'

import classNames from 'classnames'

import { Button, H1, Text } from '@sourcegraph/wildcard'

import { SetupStepsContext, StepComponentProps } from '../../../../setup-wizard/components'

import styles from './AppAllSetSetupStep.module.scss'

export const AppAllSetSetupStep: FC<StepComponentProps> = ({ className }) => {
    const { onNextStep } = useContext(SetupStepsContext)

    return (
        <div className={classNames(styles.root, className)}>
            <div className={styles.content}>
                <div className={styles.description}>
                    <H1 className={styles.descriptionHeading}>You’re all set</H1>
                    <Text className={styles.descriptionText}>
                        Start using Cody now. Access it from the Sourcegraph icon in the menu bar.
                    </Text>
                </div>

                <div className={styles.actions}>
                    <Text className={classNames(styles.descriptionText, styles.descriptionTextSmall)}>
                        <b>Tip —</b> Use Cody as an assistant while you code by arranging the Cody window side by side
                        with your favorite IDE.
                    </Text>

                    <Button size="lg" variant="primary" className={styles.actionsButton} onClick={onNextStep}>
                        Open the app →
                    </Button>
                </div>
            </div>

            <img src="https://storage.googleapis.com/sourcegraph-assets/all-set.png" alt="" className={styles.image} />
        </div>
    )
}
