import React, { useEffect, useMemo, useRef, useState } from 'react'

import { defaultKeymap, history } from '@codemirror/commands'
import { StreamLanguage, syntaxHighlighting, defaultHighlightStyle, HighlightStyle } from '@codemirror/language'
import { shell } from '@codemirror/legacy-modes/mode/shell'
import { EditorState, Extension } from '@codemirror/state'
import { EditorView, keymap } from '@codemirror/view'
import { tags } from '@lezer/highlight'
import { mdiPlus } from '@mdi/js'
import classNames from 'classnames'

import { useLazyQuery } from '@sourcegraph/http-client'
import { useCodeMirror, defaultSyntaxHighlighting } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import {
    LoadingSpinner,
    ErrorAlert,
    H2,
    Input,
    Button,
    getDefaultInputProps,
    useField,
    useForm,
    Icon,
    Form,
    Label,
    Container,
    H4,
    H3,
} from '@sourcegraph/wildcard'

import { LogOutput } from '../../../../../components/LogOutput'
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
import { InferenceForm } from '../inference-form/InferenceForm'

interface InferenceScriptPreviewFormValues {
    repository: string
}

interface InferenceScriptPreviewProps {
    script: string | null
    setTab: (index: number) => void
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
                <>
                    {/* // TODO: ul */}
                    <InferenceForm jobs={data.inferAutoIndexJobsForRepo} readOnly={false} />
                    {/* {data.inferAutoIndexJobsForRepo.map((job, index) => (
                        <IndexJobNode key={job.root} node={job} jobNumber={index + 1} />
                    ))} */}
                </>
            ) : (
                <></>
            )}
        </div>
    )
}

interface IndexJobNodeProps {
    node: AutoIndexJobDescriptionFields
    jobNumber: number
}

const IndexJobNode: React.FunctionComponent<IndexJobNodeProps> = ({ node, jobNumber }) => {
    // TODO: Check that '' === '/'
    const root = node.root === '' ? '/' : node.root

    const indexer = node.indexer?.imageName ? node.indexer.imageName : ''
    const indexerArgs = node.steps.index.indexerArgs
    const outfile = node.steps.index.outfile
    const steps = node.steps.preIndex
    const localSteps = node.steps.index.commands
    const requestedEnvVars = node.steps.index.requestedEnvVars ?? []

    return (
        <Container className={styles.job}>
            <H3 className={styles.jobHeader}>Job #{jobNumber}</H3>
            <ul className={styles.jobContent}>
                <IndexJobLabel label="Root">
                    <Input value={root} readOnly={true} className={styles.jobInput} />
                </IndexJobLabel>
                <IndexJobLabel label="Indexer">
                    <CodeMirrorCommandInput value={indexer} disabled={true} className={styles.jobInput} />
                </IndexJobLabel>
                <IndexJobLabel label="Indexer args">
                    <div className={styles.jobCommandContainer}>
                        {indexerArgs.map((arg, index) => (
                            <CodeMirrorCommandInput
                                key={index}
                                value={arg}
                                disabled={true}
                                className={styles.jobInput}
                            />
                        ))}
                        <Button variant="secondary" className="mt-2" size="sm">
                            <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                            Add arg
                        </Button>
                    </div>
                </IndexJobLabel>
                <IndexJobLabel label="Requested env vars">
                    {requestedEnvVars.length > 0 && (
                        <div className={styles.jobCommandContainer}>
                            {requestedEnvVars.map((envVar, index) => (
                                <CodeMirrorCommandInput
                                    key={index}
                                    value={envVar}
                                    disabled={true}
                                    className={styles.jobInput}
                                />
                            ))}
                            <Button variant="secondary" className="mt-2" size="sm">
                                <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                                Add env var
                            </Button>
                        </div>
                    )}
                </IndexJobLabel>
                <IndexJobLabel label="Local steps">
                    {localSteps.length > 0 && (
                        <div className={styles.jobCommandContainer}>
                            {localSteps.map((localStep, index) => (
                                <CodeMirrorCommandInput
                                    key={index}
                                    value={localStep}
                                    disabled={true}
                                    className={styles.jobInput}
                                />
                            ))}
                            <Button variant="secondary" className="mt-2" size="sm">
                                <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                                Add local step
                            </Button>
                        </div>
                    )}
                </IndexJobLabel>
                <IndexJobLabel label="Outfile">
                    {outfile ? (
                        <Input value={outfile} readOnly={true} className={styles.jobInput} />
                    ) : (
                        <Button variant="secondary" size="sm" className={styles.jobInputAction}>
                            <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                            Add outflile
                        </Button>
                    )}
                </IndexJobLabel>
                {steps.length > 0 && (
                    <Container className={styles.jobStepContainer} as="li">
                        {steps.map((step, index) => (
                            <IndexStepNode key={step.root} step={step} stepNumber={index + 1} />
                        ))}
                    </Container>
                )}
                <Button variant="secondary" className="d-block mt-2 ml-auto">
                    <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                    Add step
                </Button>
            </ul>
        </Container>
    )
}

interface IndexStepNodeProps {
    step: AutoIndexLsifPreIndexFields
    stepNumber: number
}

const IndexStepNode: React.FunctionComponent<IndexStepNodeProps> = ({ step, stepNumber }) => {
    // TODO: Check that '' === '/'
    const root = step.root === '' ? '/' : step.root

    const image = step.image ? step.image : ''
    const commands = step.commands

    return (
        <div className={styles.jobStep}>
            <H4 className={styles.jobStepHeader}>Step #{stepNumber}</H4>
            <ul className={styles.jobStepContent}>
                <IndexJobLabel label="Root">
                    <Input value={root} readOnly={true} className={styles.jobInput} />
                </IndexJobLabel>
                <IndexJobLabel label="Image">
                    <CodeMirrorCommandInput value={image} disabled={true} className={styles.jobInput} />
                </IndexJobLabel>
                <IndexJobLabel label="Commands">
                    <div className={styles.jobCommandContainer}>
                        {commands.map((command, index) => (
                            <CodeMirrorCommandInput
                                key={index}
                                value={command}
                                disabled={false}
                                className={styles.jobInput}
                            />
                        ))}
                        <Button variant="secondary" className="mt-2" size="sm">
                            <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
                            Add command
                        </Button>
                    </div>
                </IndexJobLabel>
            </ul>
        </div>
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

interface CodeMirrorCommandInputProps {
    value: string
    onChange?: (value: string) => void
    className?: string
    disabled?: boolean
}

export const shellHighlighting: Extension = [
    syntaxHighlighting(HighlightStyle.define([{ tag: [tags.keyword], class: 'hljs-keyword' }])),
    defaultSyntaxHighlighting,
]

export const CodeMirrorCommandInput: React.FunctionComponent<CodeMirrorCommandInputProps> = React.memo(
    function CodeMirrorComandInput({ value, className, disabled = false, onChange = () => {} }) {
        const containerRef = useRef<HTMLDivElement | null>(null)
        const editorRef = useRef<EditorView | null>(null)

        useCodeMirror(
            editorRef,
            containerRef,
            value,
            useMemo(
                () => [
                    EditorState.readOnly.of(disabled),
                    keymap.of(defaultKeymap),
                    history(),
                    EditorView.theme({
                        '&': {
                            flex: 1,
                            backgroundColor: 'var(--input-bg)',
                            borderRadius: 'var(--border-radius)',
                            borderColor: 'var(--border-color)',
                            marginRight: '0.5rem',
                        },
                        '&.cm-editor.cm-focused': {
                            outline: 'none',
                        },
                        '.cm-scroller': {
                            overflowX: 'hidden',
                        },
                        '.cm-content': {
                            caretColor: 'var(--search-query-text-color)',
                            fontFamily: 'var(--code-font-family)',
                            fontSize: 'var(--code-font-size)',
                        },
                        '.cm-content.focus-visible': {
                            boxShadow: 'none',
                        },
                        '.cm-line': {
                            padding: '0',
                        },
                    }),
                    StreamLanguage.define(shell),
                    shellHighlighting,
                    EditorView.updateListener.of(update => {
                        if (update.docChanged) {
                            console.log(JSON.stringify(update.state.sliceDoc()))
                            onChange(update.state.sliceDoc())
                        }
                    }),
                ],
                [disabled, onChange]
            )
        )

        return <div ref={containerRef} data-editor="codemirror6" className={classNames('form-control', className)} />
    }
)
