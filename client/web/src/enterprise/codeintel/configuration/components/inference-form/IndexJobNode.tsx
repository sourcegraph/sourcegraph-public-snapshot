import React, { useState } from 'react'

import { mdiChevronDown, mdiChevronLeft, mdiClose, mdiPlus } from '@mdi/js'
import classNames from 'classnames'
import { uniqueId } from 'lodash'

import {
    Button,
    Collapse,
    CollapseHeader,
    CollapsePanel,
    Container,
    H3,
    H4,
    Icon,
    Input,
    Tooltip,
} from '@sourcegraph/wildcard'

import { CommandInput } from './CommandInput'
import { IndexJobLabel } from './IndexJobLabel'
import type { InferenceArrayValue, InferenceFormJob, InferenceFormJobStep } from './types'
import { sanitizeIndexer, sanitizeRoot } from './util'

import styles from './IndexJobNode.module.scss'

interface IndexJobNodeProps {
    open: boolean
    job: InferenceFormJob
    jobNumber: number
    readOnly: boolean
    onChange: (name: keyof InferenceFormJob, value: unknown) => void
}

export const IndexJobNode: React.FunctionComponent<IndexJobNodeProps> = ({
    open,
    job,
    jobNumber,
    readOnly,
    onChange,
}) => {
    const [isOpened, setOpened] = useState(open)

    return (
        <Collapse isOpen={isOpened} onOpenChange={() => setOpened(!isOpened)}>
            <CollapseHeader as={H3} focusLocked={true} className={classNames(styles.jobHeader, 'mb-0')}>
                <span>
                    Job #{jobNumber}: index {sanitizeRoot(job.root)} with {sanitizeIndexer(job.indexer)}
                </span>
                <Icon aria-hidden={true} svgPath={isOpened ? mdiChevronDown : mdiChevronLeft} className="mr-1" />
            </CollapseHeader>

            <CollapsePanel>
                <ul className={classNames(styles.jobContent, 'mt-2')}>
                    <IndexJobLabel
                        label="Root"
                        tooltip="The path relative to the repository root where the indexer runs."
                    >
                        <Input
                            value={job.root}
                            onChange={event => onChange('root', event.target.value)}
                            readOnly={readOnly}
                            className={styles.jobInput}
                        />
                    </IndexJobLabel>
                    <IndexJobLabel label="Indexer" tooltip="The name of the docker image containing the indexer.">
                        <CommandInput
                            value={job.indexer}
                            onChange={value => onChange('indexer', value)}
                            readOnly={readOnly}
                            className={styles.jobInput}
                        />
                    </IndexJobLabel>
                    <IndexJobLabel label="Indexer args" tooltip="A list of arguments to pass to docker run.">
                        <IndexCommandNode
                            commands={job.indexer_args}
                            name="indexer_args"
                            actionLabel="arg"
                            readOnly={readOnly}
                            onChange={onChange}
                        />
                    </IndexJobLabel>
                    <IndexJobLabel
                        label="Requested env vars"
                        tooltip="A list of environment variables made available to the indexer."
                    >
                        <IndexCommandNode
                            commands={job.requestedEnvVars ?? []}
                            name="requestedEnvVars"
                            actionLabel="env var"
                            readOnly={readOnly}
                            onChange={onChange}
                        />
                    </IndexJobLabel>
                    <IndexJobLabel
                        label="Local steps"
                        tooltip="A command to run in the docker container to perform setup with effects outside the repository root."
                    >
                        <IndexCommandNode
                            commands={job.local_steps}
                            name="local_steps"
                            actionLabel="local step"
                            readOnly={readOnly}
                            onChange={onChange}
                        />
                    </IndexJobLabel>
                    <IndexJobLabel label="Outfile" tooltip="The path to the LSIF index relative to the index root.">
                        <Input
                            value={job.outfile}
                            onChange={event => onChange('outfile', event.target.value)}
                            readOnly={readOnly}
                            className={styles.jobInput}
                        />
                    </IndexJobLabel>
                    <Container className={styles.jobStepContainer} as="li">
                        {job.steps.map((step, index) => (
                            <div className={styles.jobStep} key={step.meta.id}>
                                <div className={styles.jobStepHeader}>
                                    <Tooltip content="A step performed before this index job. Changes are only reflected in the repository directory.">
                                        <H4 className="mb-0">Step #{index + 1}</H4>
                                    </Tooltip>
                                    {!readOnly && (
                                        <Tooltip content="Remove step">
                                            <Button
                                                variant="icon"
                                                className="ml-2 text-danger"
                                                onClick={() => {
                                                    const steps = [...job.steps]
                                                    steps.splice(index, 1)
                                                    onChange('steps', steps)
                                                }}
                                            >
                                                <Icon svgPath={mdiClose} aria-hidden={true} />
                                            </Button>
                                        </Tooltip>
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
                                className="d-block ml-auto my-3"
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
                </ul>
            </CollapsePanel>
        </Collapse>
    )
}

interface IndexStepNodeProps {
    step: InferenceFormJobStep
    readOnly: boolean
    onChange: (name: keyof InferenceFormJobStep, value: unknown) => void
}

const IndexStepNode: React.FunctionComponent<IndexStepNodeProps> = ({ step, readOnly, onChange }) => (
    <ul className={styles.jobStepContent}>
        <IndexJobLabel label="Root" tooltip="The working directory within the Docker container.">
            <Input
                value={step.root}
                onChange={event => onChange('root', event.target.value)}
                readOnly={readOnly}
                className={styles.jobInput}
            />
        </IndexJobLabel>
        <IndexJobLabel label="Image" tooltip="The docker image to run.">
            <CommandInput
                value={step.image}
                onChange={value => onChange('image', value)}
                readOnly={readOnly}
                className={styles.jobInput}
            />
        </IndexJobLabel>
        <IndexJobLabel label="Commands" tooltip="A list of arguments to pass to docker run.">
            <IndexCommandNode<keyof InferenceFormJobStep>
                commands={step.commands}
                name="commands"
                actionLabel="command"
                readOnly={readOnly}
                onChange={onChange}
            />
        </IndexJobLabel>
    </ul>
)

interface IndexCommandNodeProps<formKey = keyof InferenceFormJob> {
    name: formKey
    actionLabel: string
    commands: InferenceArrayValue[]
    onChange: (name: formKey, value: unknown) => void
    readOnly: boolean
}

const IndexCommandNode = <formKey,>({
    name,
    actionLabel,
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
                    <Tooltip content={`Remove ${actionLabel}`}>
                        <Button
                            variant="icon"
                            className="ml-2 text-danger"
                            onClick={() => {
                                const prevCommands = [...commands]
                                prevCommands.splice(index, 1)
                                onChange(name, prevCommands)
                            }}
                        >
                            <Icon svgPath={mdiClose} aria-hidden={true} />
                        </Button>
                    </Tooltip>
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
                Add {actionLabel}
            </Button>
        )}
    </div>
)
