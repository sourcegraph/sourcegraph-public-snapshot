import { FC, useContext, useState } from 'react'

import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { useQuery, gql } from '@sourcegraph/http-client'
import { Button, H1, Text, Tooltip } from '@sourcegraph/wildcard'

import { AppUserConnectDotComAccountResult, LocalRepository } from '../../../../graphql-operations'
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

    const [repositories, setRepositories] = useState<LocalRepository[]>([])

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

                <Text className={styles.descriptionText}>
                    Consider adding your most recent projects, or the repositories you edit the most. You can always add
                    more later.
                </Text>

                <Tooltip content={repositories.length === 0 ? 'Select at least one repo to continue' : undefined}>
                    <Button
                        size="lg"
                        variant="primary"
                        disabled={loading || repositories.length === 0}
                        className={styles.descriptionNext}
                        onClick={handleNext}
                    >
                        Next →
                    </Button>
                </Tooltip>
            </div>
            <div className={styles.localRepositories}>
                <LocalRepositoriesWidget
                    className={styles.localRepositoriesWidget}
                    onRepositoriesChange={setRepositories}
                >
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
