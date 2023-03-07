import { useQuery } from '@sourcegraph/http-client'
import { Button, Container, Form, H3, Input, Label } from '@sourcegraph/wildcard'
import React, { useCallback, useState } from 'react'
import AJV from 'ajv'
import addFormats from 'ajv-formats'

import {
    AutoIndexJobDescriptionFields,
    InferAutoIndexJobsForRepoResult,
    InferAutoIndexJobsForRepoVariables,
} from '../../../../../graphql-operations'
import { INFER_JOBS_SCRIPT } from './backend'

import schema from '../../schema.json'

// TODO: Own file
import styles from './InferenceScriptPreview.module.scss'

const ajv = new AJV({ strict: false })
addFormats(ajv)

// Note: This matches the auto-indexing configuration reference
// https://docs.sourcegraph.com/code_navigation/references/auto_indexing_configuration
interface InferenceFormValues {
    indexJobs: {
        root: string | null
        indexer: string | null
        indexerArgs: string[] | null
        requestedEnvVars: string[] | null
        localSteps: string[] | null
        outfile: string | null
        steps: {
            root: string | null
            image: string | null
            commands: string[] | null
        }[]
    }[]
}

interface InferenceFormProps {
    readOnly: boolean
    jobs: AutoIndexJobDescriptionFields[]
}

export const InferenceForm: React.FunctionComponent<InferenceFormProps> = ({ jobs }) => {
    const [formData, setFormData] = useState<AutoIndexJobDescriptionFields[]>(jobs)

    const handleSubmit = useCallback((event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault()

        // Validate form data against JSONSchema
        const isValid = ajv.validate(schema, formData)
        console.log(isValid)
    }, [])

    return (
        // eslint-disable-next-line react/forbid-elements
        <Form onSubmit={handleSubmit}>
            <>
                {formData.map((job, index) => (
                    <DummyIndexJobNode key={index} node={job} jobNumber={index + 1} />
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

interface DummyIndexJobNodeProps {
    node: AutoIndexJobDescriptionFields
    jobNumber: number
}

const DummyIndexJobNode: React.FunctionComponent<DummyIndexJobNodeProps> = ({ node, jobNumber }) => {
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
            </ul>
        </Container>
    )
}

// interface IndexJobNodeProps {
//     node: AutoIndexJobDescriptionFields
//     jobNumber: number
// }

// const IndexJob: React.FunctionComponent<IndexJobNodeProps> = ({ node, jobNumber }) => {
//     // TODO: Check that '' === '/'
//     const root = node.root === '' ? '/' : node.root

//     const indexer = node.indexer?.imageName ? node.indexer.imageName : ''
//     const indexerArgs = node.steps.index.indexerArgs
//     const outfile = node.steps.index.outfile
//     const steps = node.steps.preIndex
//     const localSteps = node.steps.index.commands
//     const requestedEnvVars = node.steps.index.requestedEnvVars ?? []

//     return (
//         <Container className={styles.job}>
//             <H3 className={styles.jobHeader}>Job #{jobNumber}</H3>
//             <ul className={styles.jobContent}>
//                 <IndexJobLabel label="Root">
//                     <Input value={root} readOnly={true} className={styles.jobInput} />
//                 </IndexJobLabel>
//                 <IndexJobLabel label="Indexer">
//                     <CodeMirrorCommandInput value={indexer} disabled={true} className={styles.jobInput} />
//                 </IndexJobLabel>
//                 <IndexJobLabel label="Indexer args">
//                     <div className={styles.jobCommandContainer}>
//                         {indexerArgs.map((arg, index) => (
//                             <CodeMirrorCommandInput
//                                 key={index}
//                                 value={arg}
//                                 disabled={true}
//                                 className={styles.jobInput}
//                             />
//                         ))}
//                         <Button variant="secondary" className="mt-2" size="sm">
//                             <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
//                             Add arg
//                         </Button>
//                     </div>
//                 </IndexJobLabel>
//                 <IndexJobLabel label="Requested env vars">
//                     {requestedEnvVars.length > 0 && (
//                         <div className={styles.jobCommandContainer}>
//                             {requestedEnvVars.map((envVar, index) => (
//                                 <CodeMirrorCommandInput
//                                     key={index}
//                                     value={envVar}
//                                     disabled={true}
//                                     className={styles.jobInput}
//                                 />
//                             ))}
//                             <Button variant="secondary" className="mt-2" size="sm">
//                                 <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
//                                 Add env var
//                             </Button>
//                         </div>
//                     )}
//                 </IndexJobLabel>
//                 <IndexJobLabel label="Local steps">
//                     {localSteps.length > 0 && (
//                         <div className={styles.jobCommandContainer}>
//                             {localSteps.map((localStep, index) => (
//                                 <CodeMirrorCommandInput
//                                     key={index}
//                                     value={localStep}
//                                     disabled={true}
//                                     className={styles.jobInput}
//                                 />
//                             ))}
//                             <Button variant="secondary" className="mt-2" size="sm">
//                                 <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
//                                 Add local step
//                             </Button>
//                         </div>
//                     )}
//                 </IndexJobLabel>
//                 <IndexJobLabel label="Outfile">
//                     {outfile ? (
//                         <Input value={outfile} readOnly={true} className={styles.jobInput} />
//                     ) : (
//                         <Button variant="secondary" size="sm" className={styles.jobInputAction}>
//                             <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
//                             Add outflile
//                         </Button>
//                     )}
//                 </IndexJobLabel>
//                 {steps.length > 0 && (
//                     <Container className={styles.jobStepContainer} as="li">
//                         {steps.map((step, index) => (
//                             <IndexStepNode key={step.root} step={step} stepNumber={index + 1} />
//                         ))}
//                     </Container>
//                 )}
//                 <Button variant="secondary" className="d-block mt-2 ml-auto">
//                     <Icon svgPath={mdiPlus} aria-hidden={true} className="mr-1" />
//                     Add step
//                 </Button>
//             </ul>
//         </Container>
//     )
// }
