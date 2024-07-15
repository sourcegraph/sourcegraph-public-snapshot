import React, { useState, type ReactNode } from 'react'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import {
    Alert,
    Button,
    Code,
    Container,
    ErrorAlert,
    Form,
    Input,
    InputDescription,
    Label,
    TextArea,
} from '@sourcegraph/wildcard'

import type { WorkflowInput, WorkflowUpdateInput } from '../graphql-operations'

import styles from './Form.module.scss'

export interface WorkflowFormValue
    extends Pick<WorkflowInput | WorkflowUpdateInput, 'name' | 'description' | 'templateText' | 'draft'> {}

export interface WorkflowFormProps extends TelemetryV2Props {
    initialValue?: Partial<WorkflowFormValue>
    namespaceField: ReactNode
    submitLabel: string
    onSubmit: (fields: WorkflowFormValue) => void
    otherButtons?: ReactNode
    loading: boolean
    error?: any
    flash?: ReactNode
}

const workflowNamePattern = '^[0-9A-Za-z](?:[0-9A-Za-z]|[.-](?=[0-9A-Za-z]))*-?$'

export const WorkflowForm: React.FunctionComponent<React.PropsWithChildren<WorkflowFormProps>> = ({
    initialValue,
    namespaceField,
    submitLabel,
    onSubmit,
    otherButtons,
    loading,
    error,
    flash,
}) => {
    const [value, setValue] = useState<WorkflowFormValue>(() => ({
        name: initialValue?.name || '',
        description: initialValue?.description || '',
        templateText: initialValue?.templateText || '',
        draft: initialValue?.draft || false,
    }))

    /**
     * Returns an input change handler that updates the SavedQueryFields in the component's state
     * @param key The key of saved query fields that a change of this input should update
     */
    const createInputChangeHandler =
        (key: keyof WorkflowFormValue): React.FormEventHandler<HTMLInputElement | HTMLTextAreaElement> =>
        event => {
            const { value, type } = event.currentTarget
            const checked = 'checked' in event.currentTarget ? event.currentTarget.checked : undefined
            setValue(values => ({
                ...values,
                [key]: type === 'checkbox' ? checked : value,
            }))
        }

    return (
        <Form
            onSubmit={event => {
                event.preventDefault()
                onSubmit(value)
            }}
            data-test-id="workflow-form"
            className="d-flex flex-column flex-gap-4"
        >
            <Container>
                <div className="d-flex flex-gap-4">
                    {namespaceField}
                    <div className="form-group">
                        <Label className="d-block" aria-hidden={true}>
                            &nbsp;
                        </Label>
                        <span className={styles.namespaceSlash}>/</span>
                    </div>
                    <div className="form-group">
                        <Input
                            name="description"
                            required={true}
                            value={value.name}
                            pattern={workflowNamePattern.toString()}
                            onChange={createInputChangeHandler('name')}
                            label="Workflow name"
                            autoComplete="off"
                            autoCapitalize="off"
                        />
                        <InputDescription className="mt-n1">
                            Only letters, numbers, _, and - are allowed. Example:{' '}
                            <Code>generate-typescript-e2e-tests</Code>
                        </InputDescription>
                    </div>
                </div>
                <Input
                    name="description"
                    value={value.description}
                    onChange={createInputChangeHandler('description')}
                    className="form-group"
                    autoComplete="off"
                    label="Description (optional)"
                />
                <div className="form-group">
                    <TextArea
                        name="templateText"
                        value={value.templateText}
                        onChange={createInputChangeHandler('templateText')}
                        label="Prompt template"
                        rows={10}
                        resizeable={true}
                    />
                    <InputDescription>Describe your desired output and specific requirements.</InputDescription>
                </div>
            </Container>
            <div className="d-flex flex-gap-4">
                <Button type="submit" disabled={loading} variant="primary">
                    {submitLabel}
                </Button>
                {otherButtons}
            </div>
            {flash && !loading && (
                <Alert variant="success" className="mb-0">
                    {flash}
                </Alert>
            )}
            {error && !loading && <ErrorAlert className="mb-0" error={error} />}
        </Form>
    )
}
