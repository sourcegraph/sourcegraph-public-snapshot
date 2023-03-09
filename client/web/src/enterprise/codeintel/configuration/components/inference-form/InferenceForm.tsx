import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { mdiPlus } from '@mdi/js'
import AJV from 'ajv'
import addFormats from 'ajv-formats'
import { uniqueId } from 'lodash'

import { Button, Container, Form, Icon, LoadingSpinner } from '@sourcegraph/wildcard'

import { AutoIndexJobDescriptionFields } from '../../../../../graphql-operations'
import schema from '../../schema.json'
import { ConfigurationInferButton } from '../ConfigurationInferButton'

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

    showInferButton?: boolean
    onInfer?: () => void
}

export const InferenceForm: React.FunctionComponent<InferenceFormProps> = ({
    jobs,
    readOnly,
    onSubmit,
    showInferButton,
    onInfer,
}) => {
    const firstRender = useRef(true)
    const initialFormData = useRef(autoIndexJobsToFormData({ jobs }))
    const [formData, setFormData] = useState<InferenceFormData>(initialFormData.current)
    const [loading, setLoading] = useState(false)

    // Allow the parent to update form data after the first mount
    useEffect(() => {
        if (firstRender.current) {
            firstRender.current = false
            return
        }

        setFormData(autoIndexJobsToFormData({ jobs, dirty: true }))
    }, [jobs])

    const isDirty = useMemo(() => formData.dirty, [formData])

    const handleSubmit = useCallback(
        (event: React.FormEvent<HTMLFormElement>) => {
            event.preventDefault()

            if (!onSubmit) {
                return
            }

            setLoading(true)

            const schemaCompatibleFormData = formDataToSchema(formData)

            // Validate form data against JSONSchema
            const isValid = ajv.validate(schema, schemaCompatibleFormData)

            if (isValid) {
                // eslint-disable-next-line @typescript-eslint/no-floating-promises
                onSubmit(schemaCompatibleFormData).then(() => {
                    setLoading(false)
                    // Reset dirty state
                    setFormData(previous => ({
                        ...previous,
                        dirty: false,
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
                    dirty: true,
                    index_jobs: [
                        ...previous.index_jobs.slice(0, index),
                        {
                            ...job,
                            [name]: value,
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
                dirty: true,
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
                                        dirty: true,
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
                <div className="d-flex align-items-center">
                    <Button type="submit" variant="primary" disabled={!isDirty} className="mr-2">
                        Save
                    </Button>
                    <Button
                        variant="secondary"
                        disabled={!isDirty}
                        onClick={() => setFormData(initialFormData.current)}
                    >
                        Discard changes
                    </Button>
                    {showInferButton && <ConfigurationInferButton onClick={onInfer} />}
                    {loading && <LoadingSpinner className="ml-2" />}
                </div>
            )}
        </Form>
    )
}
