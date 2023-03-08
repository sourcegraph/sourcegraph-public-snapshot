import React, { useCallback, useState } from 'react'

import AJV from 'ajv'
import addFormats from 'ajv-formats'

import { Button, Form } from '@sourcegraph/wildcard'

import { AutoIndexJobDescriptionFields } from '../../../../../graphql-operations'
import schema from '../../schema.json'

import { autoIndexJobsToFormData } from './auto-index-to-form-job'
import { formDataToSchema } from './form-data-to-schema'
import { IndexJobNode } from './IndexJobNode'
import { InferenceFormData, InferenceFormJob, SchemaCompatibleInferenceFormData } from './types'

const ajv = new AJV({ strict: false })
addFormats(ajv)

interface InferenceFormProps {
    readOnly: boolean
    jobs: AutoIndexJobDescriptionFields[]
    onSubmit?: (data: SchemaCompatibleInferenceFormData) => void
}

export const InferenceForm: React.FunctionComponent<InferenceFormProps> = ({ jobs, readOnly, onSubmit }) => {
    const [formData, setFormData] = useState<InferenceFormData>(autoIndexJobsToFormData(jobs))

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
                onSubmit(schemaCompatibleFormData)
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
                    <IndexJobNode
                        key={job.meta.id}
                        job={job}
                        jobNumber={index + 1}
                        readOnly={readOnly}
                        onChange={getChangeHandler(job.meta.id)}
                        onRemove={getRemoveHandler(job.meta.id)}
                    />
                ))}
            </>
            {!readOnly && (
                <Button type="submit" variant="primary">
                    Save
                </Button>
            )}
        </Form>
    )
}
