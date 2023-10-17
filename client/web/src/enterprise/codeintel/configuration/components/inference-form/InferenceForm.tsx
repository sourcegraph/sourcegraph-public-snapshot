import React, { useCallback, useEffect, useMemo, useState, useRef } from 'react'

import { mdiPlus, mdiClose } from '@mdi/js'
import AJV from 'ajv'
import addFormats from 'ajv-formats'
import { cloneDeep, uniqueId } from 'lodash'
import { useLocation } from 'react-router-dom'

import {
    BeforeUnloadPrompt,
    Button,
    Container,
    Form,
    Icon,
    LoadingSpinner,
    Select,
    Tooltip,
    useDeepMemo,
} from '@sourcegraph/wildcard'

import schema from '../../schema.json'
import { ConfigurationInferButton } from '../ConfigurationInferButton'

import { formDataToSchema } from './form-data-to-schema'
import { IndexJobNode } from './IndexJobNode'
import type { InferenceFormData, InferenceFormJob, SchemaCompatibleInferenceFormData } from './types'
import { sanitizeIndexer, sanitizeRoot } from './util'

import styles from './InferenceForm.module.scss'

const ajv = new AJV({ strict: false })
addFormats(ajv)

interface InferenceFormProps {
    readOnly: boolean
    initialFormData: InferenceFormData
    onSubmit?: (data: SchemaCompatibleInferenceFormData) => Promise<void>

    showInferButton?: boolean
    onInfer?: () => void
}

export const InferenceForm: React.FunctionComponent<InferenceFormProps> = ({
    initialFormData: _initialFormData,
    readOnly,
    onSubmit,
    showInferButton,
    onInfer,
}) => {
    const initialFormData = useDeepMemo(cloneDeep(_initialFormData))
    const [formData, setFormData] = useState<InferenceFormData>(initialFormData)
    const [loading, setLoading] = useState(false)

    const isDirty = useMemo(() => formData.dirty, [formData.dirty])
    const location = useLocation()

    const [openJobId, setOpenJobId] = useState<string | null>(null)
    const jobRefs = useRef(new Map())

    // Set initial scroll position
    useMemo(() => {
        if (!location.hash) {
            return
        }

        setOpenJobId(location.hash.slice(1))
    }, [location.hash])

    // Update scroll position whenever a job is forced open
    // (e.g. when a new job added)
    useEffect(() => {
        if (jobRefs.current.has(openJobId)) {
            const targetElement = jobRefs.current.get(openJobId)
            targetElement.scrollIntoView()
        }
    }, [openJobId])

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

    const addJob = useCallback(() => {
        const jobId = uniqueId()
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
                        id: jobId,
                    },
                },
            ],
        }))
        setOpenJobId(jobId)
    }, [])

    const [filter, setFilter] = useState({
        root: '',
        indexer: '',
    })
    const roots = [...new Set(formData.index_jobs.map(job => sanitizeRoot(job.root)))].sort()
    const indexers = [...new Set(formData.index_jobs.map(job => sanitizeIndexer(job.indexer)))]
    const filteredJobs = formData.index_jobs.filter(
        ({ root, indexer }: InferenceFormJob) =>
            (filter.root === '' || filter.root === sanitizeRoot(root)) &&
            (filter.indexer === '' || filter.indexer === sanitizeIndexer(indexer))
    )

    return (
        <>
            <BeforeUnloadPrompt when={isDirty} message="Discard changes?" />

            <Form onSubmit={handleSubmit}>
                <div className={styles.inputs}>
                    <span className="py-2 mr-3">
                        <Select
                            id="root"
                            label="Filter by root"
                            value={filter.root}
                            labelVariant="block"
                            onChange={event => setFilter({ ...filter, root: event.target.value })}
                            className="mb-2"
                        >
                            <option value="">All</option>
                            {roots.map(root => (
                                <option key={root} value={root}>
                                    {root}
                                </option>
                            ))}
                        </Select>
                    </span>

                    <span className="py-2">
                        <Select
                            id="indexer"
                            label="Filter by indexer"
                            value={filter.indexer}
                            labelVariant="block"
                            onChange={event => setFilter({ ...filter, indexer: event.target.value })}
                            className="mb-2"
                        >
                            <option value="">All</option>
                            {indexers.sort().map(indexer => (
                                <option key={indexer} value={indexer}>
                                    {indexer}
                                </option>
                            ))}
                        </Select>
                    </span>
                </div>

                {filteredJobs.length < formData.index_jobs.length && (
                    <div className="mb-2 px-2 text-muted">
                        {formData.index_jobs.length} total jobs, showing only {filteredJobs.length} matching jobs.
                    </div>
                )}

                {filteredJobs.map((job, index) => (
                    <div
                        key={job.meta.id}
                        className="d-flex justify-content-between align-items-baseline"
                        ref={jobElement => {
                            if (jobElement) {
                                jobRefs.current.set(job.meta.id, jobElement)
                            }
                        }}
                    >
                        <Container className={styles.job}>
                            <IndexJobNode
                                open={openJobId === job.meta.id}
                                job={job}
                                jobNumber={index + 1}
                                readOnly={readOnly}
                                onChange={getChangeHandler(job.meta.id)}
                            />
                        </Container>
                        {!readOnly && (
                            <Tooltip content="Remove job">
                                <Button
                                    variant="icon"
                                    className="ml-3 text-danger d-inline"
                                    onClick={getRemoveHandler(job.meta.id)}
                                >
                                    <Icon svgPath={mdiClose} aria-hidden={true} />
                                </Button>
                            </Tooltip>
                        )}
                    </div>
                ))}
                {!readOnly && (
                    <div className="d-flex justify-content-between">
                        <div className="d-flex align-items-center">
                            <Button type="submit" variant="primary" disabled={!isDirty} className="mr-2">
                                Save
                            </Button>
                            <Button
                                variant="secondary"
                                disabled={!isDirty}
                                onClick={() => setFormData(initialFormData)}
                            >
                                Discard changes
                            </Button>
                            {showInferButton && <ConfigurationInferButton onClick={onInfer} />}
                            {loading && <LoadingSpinner className="ml-2" />}
                        </div>
                        <Button variant="secondary" onClick={addJob} className="mr-2">
                            <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                            Add job
                        </Button>
                    </div>
                )}
            </Form>
        </>
    )
}
