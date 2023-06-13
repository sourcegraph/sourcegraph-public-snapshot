import { FC, useContext } from 'react'

import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { useQuery, gql } from '@sourcegraph/http-client'
import { Button, H1, Text } from '@sourcegraph/wildcard'

import { AppUserConnectDotComAccountResult } from '../../../../graphql-operations'
import { EnterprisePageRoutes } from '../../../../routes.constants'
import { SetupStepsContext, StepComponentProps } from '../../../../setup-wizard/components'
import { LocalRepositoriesWidget, PathsPickerActions } from '../../settings/local-repositories/LocalRepositoriesTab'

import styles from './AppLocalRepositoriesSetupStep.module.scss'

const SITE_GQL = gql`
    query SetupUserConnectDotComAccount {
        site {
            id
            appHasConnectedDotComAccount
        }
    }
`

export const AddLocalRepositoriesSetupPage: FC<StepComponentProps> = ({ className }) => {
    const navigate = useNavigate()
    const { onNextStep } = useContext(SetupStepsContext)

    const { data, loading } = useQuery<AppUserConnectDotComAccountResult, AppUserConnectDotComAccountResult>(SITE_GQL, {
        nextFetchPolicy: 'cache-first',
    })

    const handleNext = (): void => {
        if (data?.site?.appHasConnectedDotComAccount) {
            onNextStep()
            return
        }

        // Skip embeddings step if app isn't connected to the .com account
        navigate(`${EnterprisePageRoutes.AppSetup}/install-extensions`)
    }

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

                <Button
                    size="lg"
                    variant="primary"
                    disabled={loading}
                    className={styles.descriptionNext}
                    onClick={handleNext}
                >
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
