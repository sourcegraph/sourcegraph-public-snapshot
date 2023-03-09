import React, { useCallback, useMemo, useState } from 'react'

import { mdiPlus } from '@mdi/js'
import AJV from 'ajv'
import addFormats from 'ajv-formats'
import { uniqueId } from 'lodash'

import { Button, Container, Form, Icon } from '@sourcegraph/wildcard'

import { AutoIndexJobDescriptionFields } from '../../../../../graphql-operations'
import schema from '../../schema.json'

import { autoIndexJobsToFormData } from './auto-index-to-form-job'
import { formDataToSchema } from './form-data-to-schema'
import { IndexJobNode } from './IndexJobNode'
import { InferenceFormData, InferenceFormJob, SchemaCompatibleInferenceFormData } from './types'

import styles from './InferenceForm.module.scss'

const ajv = new AJV({ strict: false })
addFormats(ajv)

interface InferenceFormProps {
    readOnly: boolean
    jobs: AutoIndexJobDescriptionFields[]
    onSubmit?: (data: SchemaCompatibleInferenceFormData) => Promise<void>
}

export const InferenceForm: React.FunctionComponent<InferenceFormProps> = ({ jobs, readOnly, onSubmit }) => {
    const initialFormData = useMemo(() => autoIndexJobsToFormData(jobs), [jobs])
    const [formData, setFormData] = useState<InferenceFormData>(initialFormData)

    const isDirty = useMemo(() => formData.index_jobs.some(job => job.meta.dirty), [formData])

    const handleSubmit = useCallback(
        (event: React.FormEvent<HTMLFormElement>) => {
            event.preventDefault()

            if (!onSubmit) {
                return
            }

            const schemaCompatibleFormData = formDataToSchema(formData)

            // Validate form data against JSONSchema
            const isValid = ajv.validate(schema, schemaCompatibleFormData)

            if (isValid) {
                // eslint-disable-next-line @typescript-eslint/no-floating-promises
                onSubmit(schemaCompatibleFormData).then(() => {
                    // Reset dirty state
                    setFormData(previous => ({
                        index_jobs: previous.index_jobs.map(job => ({
                            ...job,
                            meta: {
                                ...job.meta,
                                dirty: false,
                            },
                        })),
                    }))
                })
            }
        },
        [formData, onSubmit]
    )

    const getChangeHandler = useCallback(
        (id: string) => (name: keyof InferenceFormJob, value: unknown) => {
            setFormData(previous => {
                const index = previous.index_jobs.findIndex(job => job.meta.id === id)
                const job = previous.index_jobs[index]

                return {
                    index_jobs: [
                        ...previous.index_jobs.slice(0, index),
                        {
                            ...job,
                            [name]: value,
                            meta: {
                                ...job.meta,
                                dirty: true,
                            },
                        },
                        ...previous.index_jobs.slice(index + 1),
                    ],
                }
            })
        },
        []
    )

    const getRemoveHandler = useCallback(
        (id: string) => () => {
            if (!window.confirm('Are you sure you want to remove this entire job?')) {
                return
            }

            setFormData(previous => ({
                index_jobs: previous.index_jobs.filter(job => job.meta.id !== id),
            }))
        },
        []
    )

    return (
        <Form onSubmit={handleSubmit}>
            <>
                {formData.index_jobs.map((job, index) => (
                    <Container id={job.meta.id} key={job.meta.id} className={styles.job}>
                        <IndexJobNode
                            job={job}
                            jobNumber={index + 1}
                            readOnly={readOnly}
                            onChange={getChangeHandler(job.meta.id)}
                            onRemove={getRemoveHandler(job.meta.id)}
                        />
                        {!readOnly && (
                            <Button
                                variant="secondary"
                                className="d-block mt-3 ml-auto"
                                onClick={() => {
                                    setFormData(previous => ({
                                        index_jobs: [
                                            ...previous.index_jobs,
                                            {
                                                root: '',
                                                indexer: '',
                                                indexer_args: [],
                                                requestedEnvVars: [],
                                                local_steps: [],
                                                outfile: '',
                                                steps: [],
                                                meta: {
                                                    id: uniqueId(),
                                                    dirty: true,
                                                },
                                            },
                                        ],
                                    }))
                                }}
                            >
                                <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                                Add job
                            </Button>
                        )}
                    </Container>
                ))}
            </>
            {!readOnly && (
                <>
                    <Button type="submit" variant="primary" disabled={!isDirty} className="mr-2">
                        Save
                    </Button>
                    <Button variant="secondary" disabled={!isDirty} onClick={() => setFormData(initialFormData)}>
                        Discard changes
                    </Button>
                </>
            )}
        </Form>
    )
}
