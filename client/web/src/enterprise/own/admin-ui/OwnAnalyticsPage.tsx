import { FC, useState } from 'react'

import { useQuery } from '@sourcegraph/http-client'
import { Container, ErrorAlert, LoadingSpinner, PageHeader, Text } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../components/PageTitle'
import { GetOwnSignalConfigurationsResult, OwnSignalConfig } from '../../../graphql-operations'

import { GET_OWN_JOB_CONFIGURATIONS } from './query'

import styles from './own-status-page-styles.module.scss'

export const OwnAnalyticsPage: FC = () => {
    const [localData, setLocalData] = useState<OwnSignalConfig[]>([])
    const [saveError] = useState<Error | null>()

    const { loading, error } = useQuery<GetOwnSignalConfigurationsResult>(GET_OWN_JOB_CONFIGURATIONS, {
        onCompleted: data => {
            setLocalData(data.ownSignalConfigurations)
        },
    })

    // const [saveConfigs, { loading: loadingSaveConfigs }] = useMutation<
    //     UpdateSignalConfigsResult,
    //     UpdateSignalConfigsVariables
    // >(UPDATE_SIGNAL_CONFIGURATIONS, {})

    // function onUpdateJob(index: number, newJob: OwnSignalConfig): void {
    //     setHasLocalChanges(true)
    //     const newData = localData.map((job: OwnSignalConfig, ind: number) => {
    //         if (ind === index) {
    //             return newJob
    //         }
    //         return job
    //     })
    //     setLocalData(newData)
    // }

    return (
        <div>
            <span className={styles.topHeader}>
                <div>
                    <PageTitle title="Own Analytics" />
                    <PageHeader
                        headingElement="h2"
                        path={[{ text: 'Own Analytics' }]}
                        description="TODO"
                        className="mb-3"
                    />
                    {saveError && <ErrorAlert error={saveError} />}
                </div>
            </span>

            <Container className={styles.root}>
                {loading && <LoadingSpinner />}
                {error && <ErrorAlert prefix="Error fetching Own signal configurations" error={error} />}
                {!loading && localData && !error && <Text>Foo bar</Text>}
            </Container>
        </div>
    )
}
