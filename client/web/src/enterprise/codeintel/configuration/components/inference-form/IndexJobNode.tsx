import React from 'react'

import { mdiClose, mdiPlus } from '@mdi/js'
import { uniqueId } from 'lodash'

import { Button, Container, H3, H4, Icon, Input } from '@sourcegraph/wildcard'

import { CommandInput } from './CommandInput'
import { IndexJobLabel } from './IndexJobLabel'
import { InferenceArrayValue, InferenceFormJob, InferenceFormJobStep } from './types'

import styles from './IndexJobNode.module.scss'

interface IndexJobNodeProps {
    job: InferenceFormJob
    jobNumber: number
    readOnly: boolean
    onChange: (name: keyof InferenceFormJob, value: unknown) => void
    onRemove: () => void
}

export const IndexJobNode: React.FunctionComponent<IndexJobNodeProps> = ({
    job,
    jobNumber,
    readOnly,
    onChange,
    onRemove,
}) => (
    <>
        <div className={styles.jobHeader}>
            <H3 className="mb-0">Job #{jobNumber}</H3>
            {!readOnly && (
                <Button variant="icon" className="ml-2 text-danger" aria-label="Remove" onClick={onRemove}>
                    <Icon svgPath={mdiClose} aria-hidden={true} />
                </Button>
            )}
        </div>
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
                <IndexCommandNode
                    commands={job.indexer_args}
                    name="indexer_args"
                    addLabel="arg"
                    readOnly={readOnly}
                    onChange={onChange}
                />
            </IndexJobLabel>
            <IndexJobLabel label="Requested env vars">
                <IndexCommandNode
                    commands={job.requestedEnvVars ?? []}
                    name="requestedEnvVars"
                    addLabel="env var"
                    readOnly={readOnly}
                    onChange={onChange}
                />
            </IndexJobLabel>
            <IndexJobLabel label="Local steps">
                <IndexCommandNode
                    commands={job.local_steps}
                    name="local_steps"
                    addLabel="local step"
                    readOnly={readOnly}
                    onChange={onChange}
                />
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
                        <div className={styles.jobStep} key={step.meta.id}>
                            <div className={styles.jobStepHeader}>
                                <H4 className="mb-0">Step #{index + 1}</H4>
                                {!readOnly && (
                                    <Button
                                        variant="icon"
                                        className="ml-2 text-danger"
                                        aria-label="Remove"
                                        onClick={() => {
                                            const steps = [...job.steps]
                                            steps.splice(index, 1)
                                            onChange('steps', steps)
                                        }}
                                    >
                                        <Icon svgPath={mdiClose} aria-hidden={true} />
                                    </Button>
                                )}
                            </div>
                            <IndexStepNode
                                step={step}
                                readOnly={readOnly}
                                onChange={(name, value) => {
                                    const steps = [...job.steps]
                                    steps[index] = { ...steps[index], [name]: value }
                                    onChange('steps', steps)
                                }}
                            />
                        </div>
                    ))}
                    {!readOnly && (
                        <Button
                            variant="secondary"
                            className="d-block mb-3 ml-auto"
                            onClick={() => {
                                onChange('steps', [
                                    ...job.steps,
                                    { root: '', image: '', commands: [], meta: { id: uniqueId() } },
                                ])
                            }}
                        >
                            <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                            Add step
                        </Button>
                    )}
                </Container>
            )}
        </ul>
    </>
)

interface IndexStepNodeProps {
    step: InferenceFormJobStep
    readOnly: boolean
    onChange: (name: keyof InferenceFormJobStep, value: unknown) => void
}

const IndexStepNode: React.FunctionComponent<IndexStepNodeProps> = ({ step, readOnly, onChange }) => (
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
            <IndexCommandNode<keyof InferenceFormJobStep>
                commands={step.commands}
                name="commands"
                addLabel="command"
                readOnly={readOnly}
                onChange={onChange}
            />
        </IndexJobLabel>
    </ul>
)

interface IndexCommandNodeProps<formKey = keyof InferenceFormJob> {
    name: formKey
    addLabel: string
    commands: InferenceArrayValue[]
    onChange: (name: formKey, value: unknown) => void
    readOnly: boolean
}

const IndexCommandNode = <formKey,>({
    name,
    addLabel,
    commands,
    onChange,
    readOnly,
}: IndexCommandNodeProps<formKey>): JSX.Element | null => (
    <div className={styles.jobCommandContainer}>
        {commands.map((command, index) => (
            <div className={styles.jobCommand} key={command.meta.id}>
                <CommandInput
                    value={command.value}
                    onChange={value => {
                        const prevCommands = [...commands]
                        prevCommands[index].value = value
                        onChange(name, prevCommands)
                    }}
                    readOnly={readOnly}
                    className={styles.jobInput}
                />
                {!readOnly && (
                    <Button
                        variant="icon"
                        className="ml-2 text-danger"
                        aria-label="Remove"
                        onClick={() => {
                            const prevCommands = [...commands]
                            prevCommands.splice(index, 1)
                            onChange(name, prevCommands)
                        }}
                    >
                        <Icon svgPath={mdiClose} aria-hidden={true} />
                    </Button>
                )}
            </div>
        ))}
        {!readOnly && (
            <Button
                variant="secondary"
                size="sm"
                onClick={() => {
                    onChange(name, [
                        ...commands,
                        {
                            value: '',
                            meta: {
                                id: uniqueId(),
                            },
                        },
                    ])
                }}
            >
                <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                Add {addLabel}
            </Button>
        )}
    </div>
)
