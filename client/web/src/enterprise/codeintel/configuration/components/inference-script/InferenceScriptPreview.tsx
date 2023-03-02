import React from 'react'

import { useLazyQuery } from '@sourcegraph/http-client'
import {
    LoadingSpinner,
    ErrorAlert,
    H2,
    Input,
    Button,
    getDefaultInputProps,
    useField,
    useForm,
    Form,
    Label,
    Container,
    H4,
    H3,
} from '@sourcegraph/wildcard'

import {
    GetRepoIdResult,
    GetRepoIdVariables,
    InferAutoIndexJobsForRepoResult,
    InferAutoIndexJobsForRepoVariables,
    AutoIndexJobDescriptionFields,
    AutoIndexLsifPreIndexFields,
} from '../../../../../graphql-operations'
import { RepositoryField } from '../../../../insights/components'

import { GET_REPO_ID, INFER_JOBS_SCRIPT } from './backend'

import styles from './InferenceScriptPreview.module.scss'

interface InferenceScriptPreviewFormValues {
    repository: string
}

interface InferenceScriptPreviewProps {
    script: string | null
}

export const InferenceScriptPreview: React.FunctionComponent<InferenceScriptPreviewProps> = ({ script }) => {
    const [getRepoId] = useLazyQuery<GetRepoIdResult, GetRepoIdVariables>(GET_REPO_ID, {})
    const [inferJobs, { data, loading, error, called }] = useLazyQuery<
        InferAutoIndexJobsForRepoResult,
        InferAutoIndexJobsForRepoVariables
    >(INFER_JOBS_SCRIPT, {
        nextFetchPolicy: 'cache-first',
    })

    const form = useForm<InferenceScriptPreviewFormValues>({
        initialValues: { repository: '' },
        onSubmit: async ({ repository }) => {
            const { data } = await getRepoId({ variables: { name: repository } })
            const id = data?.repository?.id

            if (id) {
                return inferJobs({ variables: { repository: id, script, rev: null } })
            }
        },
    })

    const repository = useField({
        name: 'repository',
        formApi: form.formAPI,
    })

    return (
        <div className={styles.container}>
            {called && loading && <LoadingSpinner className="d-block mx-auto mt-3" />}
            {called && error && <ErrorAlert error={error} />}
            {!called && (
                <Form className={styles.actionContainer} ref={form.ref} noValidate={true} onSubmit={form.handleSubmit}>
                    <Label id="preview-label">Run your script against a repository</Label>
                    <div className="d-flex align-items-center">
                        <Input
                            as={RepositoryField}
                            required={true}
                            autoFocus={true}
                            aria-label="Repository"
                            placeholder="Example: github.com/sourcegraph/sourcegraph"
                            {...getDefaultInputProps(repository)}
                            className={styles.actionInput}
                        />

                        <Button variant="success" type="submit">
                            Preview results
                        </Button>
                    </div>
                </Form>
            )}
            {data && (
                <div className={styles.resultsContainer}>
                    {data.inferAutoIndexJobsForRepo.map((job, index) => (
                        <IndexJobNode key={job.root} node={job} jobNumber={index + 1} />
                    ))}
                </div>
            )}
        </div>
    )
}

interface IndexJobNodeProps {
    node: AutoIndexJobDescriptionFields
    jobNumber: number
}

const IndexJobNode: React.FunctionComponent<IndexJobNodeProps> = ({ node, jobNumber }) => {
    // Fields
    const root = node.root
    const indexer = node.indexer ? node.indexer.name : ''
    const indexerArgs = node.steps.index.indexerArgs.length > 0 ? node.steps.index.indexerArgs.join(' ') : ''
    const outfile = node.steps.index.outfile ? node.steps.index.outfile : ''
    const steps = node.steps.preIndex

    // TODO: No localSteps?
    // const localSteps = node.steps.setup

    // TODO: No requestedEnvVars?
    // const requestedEnvVars = node.steps.setup

    return (
        <Container className="my-3">
            <H3>Index job {jobNumber}</H3>
            <IndexJobField label="Root" value={root} />
            <IndexJobField label="Indexer" value={indexer} />
            <IndexJobField label="Indexer args" value={indexerArgs} />
            <IndexJobField label="Outfile" value={outfile} />
            {steps.map((step, index) => (
                <IndexStepNode key={step.root} step={step} stepNumber={index + 1} />
            ))}
        </Container>
    )
}

interface IndexStepNodeProps {
    step: AutoIndexLsifPreIndexFields
    stepNumber: number
}

const IndexStepNode: React.FunctionComponent<IndexStepNodeProps> = ({ step, stepNumber }) => {
    const root = step.root
    const indexer = step.image ? step.image : ''
    const commands = step.commands.length > 0 ? step.commands.join(' ') : ''

    return (
        <Container className={styles.step}>
            <H4>Step {stepNumber}</H4>
            <IndexJobField label="Root" value={root} />
            <IndexJobField label="Indexer" value={indexer} />
            <IndexJobField label="Commands" value={commands} />
        </Container>
    )
}

interface IndexJobFieldProps {
    label: string
    value: string
}

const IndexJobField: React.FunctionComponent<IndexJobFieldProps> = ({ label, value }) => {
    return (
        <Label className={styles.label}>
            {label}:
            <Input value={value} disabled={true} className="ml-2" />
        </Label>
    )
}
