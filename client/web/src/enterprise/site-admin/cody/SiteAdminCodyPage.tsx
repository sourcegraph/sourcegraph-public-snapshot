import { FC, useEffect, useCallback, useMemo } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiMapSearch } from '@mdi/js'
import { Subject } from 'rxjs'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Button,
    Container,
    ErrorAlert,
    Form,
    getDefaultInputProps,
    H3,
    Icon,
    Label,
    PageHeader,
    Text,
    useField,
    useForm,
    Validator,
} from '@sourcegraph/wildcard'

import { FilteredConnection, FilteredConnectionQueryArguments } from '../../../components/FilteredConnection'
import { PageTitle } from '../../../components/PageTitle'
import { RepoEmbeddingJobConnectionFields, RepoEmbeddingJobFields } from '../../../graphql-operations'
import { RepositoriesField } from '../../insights/components'

import { repoEmbeddingJobs, useScheduleContextDetectionEmbeddingJob, useScheduleRepoEmbeddingJobs } from './backend'
import { RepoEmbeddingJobNode } from './RepoEmbeddingJobNode'

import styles from './SiteAdminCodyPage.module.scss'

export interface SiteAdminCodyPageProps extends TelemetryProps {}

interface RepoEmbeddingJobsFormValues {
    repositories: string[]
}

const INITIAL_REPOSITORIES = { repositories: [] }

const repositoriesValidator: Validator<string[]> = value => {
    if (value !== undefined && value.length === 0) {
        return 'Repositories is a required field.'
    }
    return
}

export const SiteAdminCodyPage: FC<SiteAdminCodyPageProps> = ({ telemetryService }) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminCodyPage')
    }, [telemetryService])
    const refresh = useMemo(() => new Subject<undefined>(), [])

    const [scheduleRepoEmbeddingJobs, { loading: repoEmbeddingJobsLoading, error: repoEmbeddingJobsError }] =
        useScheduleRepoEmbeddingJobs()

    const [
        scheduleContextDetectionEmbeddingJob,
        { loading: contextDetectionEmbeddingJobLoading, error: contextDetectionEmbeddingJobError },
    ] = useScheduleContextDetectionEmbeddingJob()

    const onSubmit = useCallback(
        async (repoNames: string[]) => {
            await Promise.all([
                scheduleContextDetectionEmbeddingJob(),
                scheduleRepoEmbeddingJobs({ variables: { repoNames } }),
            ])
            refresh.next()
        },
        [scheduleContextDetectionEmbeddingJob, scheduleRepoEmbeddingJobs]
    )

    const form = useForm<RepoEmbeddingJobsFormValues>({
        initialValues: INITIAL_REPOSITORIES,
        touched: false,
        onSubmit: values => onSubmit(values.repositories),
    })

    const repositories = useField({
        name: 'repositories',
        formApi: form.formAPI,
        validators: { sync: repositoriesValidator },
    })

    const apolloClient = useApolloClient()

    const queryConnection = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            return repoEmbeddingJobs(args, apolloClient)
        },
        [repoEmbeddingJobs, apolloClient]
    )

    return (
        <>
            <PageTitle title="Cody" />
            <PageHeader path={[{ text: 'Cody' }]} className="mb-3" headingElement="h2" />
            <Container>
                <H3>Schedule repositories for embedding</H3>
                <Form ref={form.ref} noValidate={true} onSubmit={form.handleSubmit}>
                    <Label htmlFor="repositories-id" className="mt-1">
                        Repositories
                    </Label>
                    <div className="d-flex">
                        <RepositoriesField
                            id="repositories-id"
                            description="Schedule repositories for embedding at latest revision on the default branch."
                            placeholder="Search repositories..."
                            className="flex-1 mr-2"
                            {...getDefaultInputProps(repositories)}
                        />
                        <div>
                            <Button
                                type="submit"
                                variant="secondary"
                                className={styles.scheduleButton}
                                disabled={repoEmbeddingJobsLoading || contextDetectionEmbeddingJobLoading}
                            >
                                {repoEmbeddingJobsLoading || contextDetectionEmbeddingJobLoading
                                    ? 'Scheduling...'
                                    : 'Schedule'}
                            </Button>
                        </div>
                    </div>
                </Form>
                {(repoEmbeddingJobsError || contextDetectionEmbeddingJobError) && (
                    <div className="mt-1">
                        <ErrorAlert
                            prefix="Error scheduling embedding jobs"
                            error={repoEmbeddingJobsError || contextDetectionEmbeddingJobError}
                        />
                    </div>
                )}
                <H3 className="mt-3">Repository embedding jobs</H3>
                <FilteredConnection<
                    RepoEmbeddingJobFields,
                    RepoEmbeddingJobFields,
                    {},
                    RepoEmbeddingJobConnectionFields
                >
                    className="mb-0 mt-1"
                    listComponent="div"
                    inputClassName="ml-2 flex-1"
                    listClassName="mb-3"
                    noun="repository embedding job"
                    pluralNoun="repository embedding jobs"
                    defaultFirst={10}
                    queryConnection={queryConnection}
                    nodeComponent={RepoEmbeddingJobNode}
                    hideSearch={true}
                    updates={refresh}
                    emptyElement={<EmptyIndex />}
                    withCenteredSummary={true}
                />
            </Container>
        </>
    )
}

const EmptyIndex: React.FunctionComponent<{}> = () => (
    <Text alignment="center" className="text-muted w-100 mb-0 mt-1">
        <Icon className="mb-2" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <br />
        No repository embedding jobs have been created so far.
    </Text>
)
