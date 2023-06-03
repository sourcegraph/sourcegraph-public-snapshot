import { FC, useContext } from 'react'

import classNames from 'classnames'

import { Button, H1, Text } from '@sourcegraph/wildcard'

import { SetupStepsContext, StepComponentProps } from '../../../../setup-wizard/components'
import { LocalRepositoriesWidget, PathsPickerActions } from '../../settings/local-repositories/LocalRepositoriesTab'

import styles from './AppLocalRepositoriesSetupStep.module.scss'

export const AddLocalRepositoriesSetupPage: FC<StepComponentProps> = ({ className }) => {
    const { onNextStep } = useContext(SetupStepsContext)

    return (
        <div className={classNames(className, styles.root)}>
            <div className={styles.description}>
                <H1 className={styles.descriptionHeading}>Add your projects</H1>

                <Text className={styles.descriptionText}>
                    Choose the local repositories you’d like to add to the app.
                </Text>

                <Text className={classNames(styles.descriptionText, styles.descriptionTextSmall)}>
                    Consider adding your most recent projects, or the repositories you edit the most. You can always add
                    more later.
                </Text>

                <Button variant="primary" size="lg" className={styles.descriptionNext} onClick={onNextStep}>
                    Next →
                </Button>
            </div>
            <div className={styles.localRepositories}>
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
