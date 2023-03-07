import { useQuery } from '@sourcegraph/http-client'
import { Button, Container, Form, H3, H4, Icon, Input, Label } from '@sourcegraph/wildcard'
import React, { useCallback, useState } from 'react'
import AJV from 'ajv'
import addFormats from 'ajv-formats'

import {
    AutoIndexJobDescriptionFields,
    InferAutoIndexJobsForRepoResult,
    InferAutoIndexJobsForRepoVariables,
} from '../../../../../graphql-operations'
import { INFER_JOBS_SCRIPT } from '../inference-script/backend'
import { InferenceFormData, InferenceFormJob, InferenceFormJobStep } from './types'
import { autoIndexJobsToFormData } from './auto-index-to-form-job'
import schema from '../../schema.json'

// TODO: Own file
import styles from '../inference-script/InferenceScriptPreview.module.scss'
import { CommandInput } from './CommandInput'
import { mdiClose, mdiPlus } from '@mdi/js'

const ajv = new AJV({ strict: false })
addFormats(ajv)

interface InferenceFormProps {
    readOnly: boolean
    jobs: AutoIndexJobDescriptionFields[]
}

export const InferenceForm: React.FunctionComponent<InferenceFormProps> = ({ jobs }) => {
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
        (comparisonKey: string) => (name: keyof InferenceFormJob, value: unknown) => {
            setFormData(previous => {
                const index = previous.index_jobs.findIndex(job => job.meta.comparisonKey === comparisonKey)
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

    return (
        // eslint-disable-next-line react/forbid-elements
        <Form onSubmit={handleSubmit}>
            <>
                {formData.index_jobs.map((job, index) => (
                    <IndexJobNode
                        key={job.meta.comparisonKey}
                        job={job}
                        jobNumber={index + 1}
                        readOnly={false}
                        onChange={getChangeHandler(job.meta.comparisonKey)}
                    />
                ))}
            </>
            <Button type="submit" variant="primary">
                Save
            </Button>
        </Form>
    )
}

interface IndexJobFieldProps {
    label: string
}

const IndexJobLabel: React.FunctionComponent<React.PropsWithChildren<IndexJobFieldProps>> = ({ label, children }) => (
    <>
        <li className={styles.jobField}>
            <Label className={styles.jobLabel}>{label}:</Label>
            {children}
        </li>
    </>
)

interface IndexJobNodeProps {
    job: InferenceFormJob
    jobNumber: number
    readOnly: boolean
    onChange: (name: keyof InferenceFormJob, value: unknown) => void
}

const IndexJobNode: React.FunctionComponent<IndexJobNodeProps> = ({ job, jobNumber, readOnly, onChange }) => (
    <Container className={styles.job}>
        <H3 className={styles.jobHeader}>Job #{jobNumber}</H3>
        <ul className={styles.jobContent}>
            <IndexJobLabel label="Root">
                <Input
                    value={job.root}
                    onChange={event => onChange('root', event.target.value)}
                    readOnly={readOnly}
                    className={styles.jobInput}
                />
            </IndexJobLabel>
            <IndexJobLabel label="Indexer">
                <CommandInput
                    value={job.indexer}
                    onChange={value => onChange('indexer', value)}
                    readOnly={readOnly}
                    className={styles.jobInput}
                />
            </IndexJobLabel>
            <IndexJobLabel label="Indexer args">
                <div className={styles.jobCommandContainer}>
                    {job.indexer_args.map((arg, index) => (
                        <div className="d-flex mb-2" key={`job-${jobNumber}-indexer-arg-${index}`}>
                            <CommandInput
                                value={arg}
                                onChange={value => {
                                    const args = [...job.indexer_args]
                                    args[index] = value
                                    onChange('indexer_args', args)
                                }}
                                readOnly={readOnly}
                                className={styles.jobInput}
                            />
                            <Button
                                variant="icon"
                                className="ml-2"
                                size="sm"
                                aria-label="Remove"
                                onClick={() => {
                                    const args = [...job.indexer_args]
                                    args.splice(index, 1)
                                    onChange('indexer_args', args)
                                }}
                            >
                                <Icon svgPath={mdiClose} aria-hidden={true} />
                            </Button>
                        </div>
                    ))}
                    <Button
                        variant="secondary"
                        className="mt-2"
                        size="sm"
                        onClick={() => {
                            onChange('indexer_args', [...job.indexer_args, ''])
                        }}
                    >
                        <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                        Add arg
                    </Button>
                </div>
            </IndexJobLabel>
            <IndexJobLabel label="Requested env vars">
                <div className={styles.jobCommandContainer}>
                    {(job.requestedEnvVars ?? []).map((envVar, index) => (
                        <CommandInput
                            key={`job-${jobNumber}-env-var-${index}`}
                            value={envVar}
                            onChange={value => {
                                const envVars = [...(job.requestedEnvVars ?? [])]
                                envVars[index] = value
                                onChange('requestedEnvVars', envVars)
                            }}
                            readOnly={readOnly}
                            className={styles.jobInput}
                        />
                    ))}
                    <Button
                        variant="secondary"
                        className="mt-2"
                        size="sm"
                        onClick={() => {
                            onChange('requestedEnvVars', [...(job.requestedEnvVars ?? []), ''])
                        }}
                    >
                        <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                        Add env var
                    </Button>
                </div>
            </IndexJobLabel>
            <IndexJobLabel label="Local steps">
                <div className={styles.jobCommandContainer}>
                    {job.local_steps.map((localStep, index) => (
                        <CommandInput
                            key={`job-${jobNumber}-local-step-${index}`}
                            value={localStep}
                            onChange={value => {
                                const localSteps = [...job.local_steps]
                                localSteps[index] = value
                                onChange('local_steps', localSteps)
                            }}
                            readOnly={readOnly}
                            className={styles.jobInput}
                        />
                    ))}
                    <Button
                        variant="secondary"
                        className="mt-2"
                        size="sm"
                        onClick={() => {
                            onChange('local_steps', [...job.local_steps, ''])
                        }}
                    >
                        <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                        Add local step
                    </Button>
                </div>
            </IndexJobLabel>
            <IndexJobLabel label="Outfile">
                <Input
                    value={job.outfile}
                    onChange={event => onChange('outfile', event.target.value)}
                    readOnly={readOnly}
                    className={styles.jobInput}
                />
            </IndexJobLabel>
            {job.steps.length > 0 && (
                <Container className={styles.jobStepContainer} as="li">
                    {job.steps.map((step, index) => (
                        <IndexStepNode
                            key={`job-${jobNumber}-step-${index}`}
                            step={step}
                            jobNumber={jobNumber}
                            stepNumber={index + 1}
                            readOnly={readOnly}
                            onChange={(name, value) => {
                                const steps = [...job.steps]
                                steps[index] = { ...steps[index], [name]: value }
                                onChange('steps', steps)
                            }}
                        />
                    ))}
                </Container>
            )}
            <Button
                variant="secondary"
                className="d-block mt-2 ml-auto"
                onClick={() => {
                    onChange('steps', [...job.steps, { root: '', image: '', commands: [] }])
                }}
            >
                <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                Add step
            </Button>
        </ul>
    </Container>
)

interface IndexStepNodeProps {
    step: InferenceFormJobStep
    jobNumber: number
    stepNumber: number
    readOnly: boolean
    onChange: (name: keyof InferenceFormJobStep, value: unknown) => void
}

const IndexStepNode: React.FunctionComponent<IndexStepNodeProps> = ({
    step,
    jobNumber,
    stepNumber,
    readOnly,
    onChange,
}) => (
    <div className={styles.jobStep}>
        <H4 className={styles.jobStepHeader}>Step #{stepNumber}</H4>
        <ul className={styles.jobStepContent}>
            <IndexJobLabel label="Root">
                <Input
                    value={step.root}
                    onChange={event => onChange('root', event.target.value)}
                    readOnly={readOnly}
                    className={styles.jobInput}
                />
            </IndexJobLabel>
            <IndexJobLabel label="Image">
                <CommandInput
                    value={step.image}
                    onChange={value => onChange('image', value)}
                    readOnly={readOnly}
                    className={styles.jobInput}
                />
            </IndexJobLabel>
            <IndexJobLabel label="Commands">
                <div className={styles.jobCommandContainer}>
                    {step.commands.map((command, index) => (
                        <CommandInput
                            key={`job-${jobNumber}-commands-${index}`}
                            value={command}
                            onChange={value => {
                                const commands = [...step.commands]
                                commands[index] = value
                                onChange('commands', commands)
                            }}
                            readOnly={readOnly}
                            className={styles.jobInput}
                        />
                    ))}
                    <Button
                        variant="secondary"
                        className="mt-2"
                        size="sm"
                        onClick={() => {
                            onChange('commands', [...step.commands, ''])
                        }}
                    >
                        <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                        Add command
                    </Button>
                </div>
            </IndexJobLabel>
        </ul>
    </div>
)
