import React, { useCallback, useState } from 'react'

import AJV from 'ajv'
import addFormats from 'ajv-formats'

import { Button, Form } from '@sourcegraph/wildcard'

import { AutoIndexJobDescriptionFields } from '../../../../../graphql-operations'
import schema from '../../schema.json'

import { autoIndexJobsToFormData } from './auto-index-to-form-job'
import { IndexJobNode } from './IndexJobNode'
import { InferenceFormData, InferenceFormJob } from './types'

const ajv = new AJV({ strict: false })
addFormats(ajv)

interface InferenceFormProps {
    readOnly: boolean
    jobs: AutoIndexJobDescriptionFields[]
}

export const InferenceForm: React.FunctionComponent<InferenceFormProps> = ({ jobs, readOnly }) => {
    const [formData, setFormData] = useState<InferenceFormData>(autoIndexJobsToFormData(jobs))

    const handleSubmit = useCallback((event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault()

        const formDataWithoutMeta = {
            index_jobs: formData.index_jobs.map(job => {
                const { meta, ...rest } = job
                return rest
            }),
        }

        // Validate form data against JSONSchema
        const isValid = ajv.validate(schema, formDataWithoutMeta)
        console.log(isValid)
    }, [])

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
