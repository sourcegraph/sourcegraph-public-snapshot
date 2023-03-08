import React from 'react'

import { useLazyQuery } from '@sourcegraph/http-client'
import {
    LoadingSpinner,
    ErrorAlert,
    Input,
    Button,
    getDefaultInputProps,
    useField,
    useForm,
    Form,
    Label,
} from '@sourcegraph/wildcard'

import {
    GetRepoIdResult,
    GetRepoIdVariables,
    InferAutoIndexJobsForRepoResult,
    InferAutoIndexJobsForRepoVariables,
} from '../../../../../graphql-operations'
import { RepositoryField } from '../../../../insights/components'
import { InferenceForm } from '../inference-form/InferenceForm'

import { GET_REPO_ID, INFER_JOBS_SCRIPT } from './backend'

import styles from './InferenceScriptPreview.module.scss'

interface InferenceScriptPreviewFormValues {
    repository: string
}

interface InferenceScriptPreviewProps {
    script: string | null
    setTab: (index: number) => void
}

export const InferenceScriptPreview: React.FunctionComponent<InferenceScriptPreviewProps> = ({ script }) => {
    const [getRepoId] = useLazyQuery<GetRepoIdResult, GetRepoIdVariables>(GET_REPO_ID, {})
    const [inferJobs, { data, loading, error }] = useLazyQuery<
        InferAutoIndexJobsForRepoResult,
        InferAutoIndexJobsForRepoVariables
    >(INFER_JOBS_SCRIPT, {})

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
            {loading ? (
                <LoadingSpinner className="d-block mx-auto mt-3" />
            ) : error ? (
                <ErrorAlert error={error} />
            ) : data ? (
                <InferenceForm jobs={data.inferAutoIndexJobsForRepo} readOnly={true} />
            ) : (
                <></>
            )}
        </div>
    )
}
