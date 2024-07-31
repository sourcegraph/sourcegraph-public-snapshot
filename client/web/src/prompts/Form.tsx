import React, { useMemo, useState, type ComponentProps, type FunctionComponent, type ReactNode } from 'react'

import {
    ClientStateContextProvider,
    Command,
    CommandEmpty,
    CommandGroup,
    CommandInput,
    CommandItem,
    CommandList,
    CommandLoading,
    CommandSeparator,
    ExtensionAPIProviderForTestsOnly,
    PromptEditor,
    PromptEditorConfigProvider,
    type PromptEditorConfig,
} from '@sourcegraph/prompt-editor'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import {
    Alert,
    Button,
    Checkbox,
    Code,
    Container,
    ErrorAlert,
    Form,
    Input,
    InputDescription,
    Label,
    TextArea,
} from '@sourcegraph/wildcard'

import { PatternConstrainedInput } from '../components/PatternConstrainedInput'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import type { PromptInput, PromptUpdateInput } from '../graphql-operations'

import styles from './Form.module.scss'

export type PromptFormValue = Pick<PromptInput | PromptUpdateInput, 'name' | 'description' | 'definitionText' | 'draft'>

export interface PromptFormProps extends TelemetryV2Props {
    initialValue?: Partial<PromptFormValue>
    namespaceField: ReactNode
    submitLabel: string
    onSubmit: (fields: PromptFormValue) => void
    otherButtons?: ReactNode
    loading: boolean
    error?: any
    flash?: ReactNode
    afterFields?: ReactNode
}

const promptNamePattern = '[0-9A-Za-z](?:[0-9A-Za-z]|[.\\-](?=[0-9A-Za-z]))*-?'

export const PromptForm: React.FunctionComponent<React.PropsWithChildren<PromptFormProps>> = ({
    initialValue,
    namespaceField,
    submitLabel,
    onSubmit,
    otherButtons,
    loading,
    error,
    flash,
    afterFields,
}) => {
    const [richPromptEditor] = useFeatureFlag('rich-prompt-editor')

    const [value, setValue] = useState<PromptFormValue>(() => ({
        name: initialValue?.name ?? '',
        description: initialValue?.description ?? '',
        definitionText: initialValue?.definitionText ?? '',
        draft: initialValue?.draft ?? true,
    }))

    /**
     * Returns an input change handler that updates the SavedQueryFields in the component's state
     * @param key The key of saved query fields that a change of this input should update
     */
    const createInputChangeHandler =
        (key: keyof PromptFormValue): React.FormEventHandler<HTMLInputElement | HTMLTextAreaElement> =>
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
            data-test-id="prompt-form"
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
                        <PatternConstrainedInput
                            label="Prompt name"
                            name="name"
                            value={value.name}
                            replaceSpaces={true}
                            pattern={promptNamePattern.toString()}
                            onChange={value => setValue(prev => ({ ...prev, name: value }))}
                            required={true}
                            autoComplete="off"
                            autoCapitalize="off"
                        />
                        <InputDescription className="mt-n1">
                            Only letters, numbers, and <Code>_.-</Code> are allowed. Example:{' '}
                            <Code>generate-unit-test</Code>
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
                    <PromptTextField
                        name="definitionText"
                        value={value.definitionText}
                        onChange={createInputChangeHandler('definitionText')}
                        label="Prompt template"
                        richPromptEditor={richPromptEditor}
                    />
                    <InputDescription>Describe your desired output and specific requirements.</InputDescription>
                </div>
                <div className="form-group d-flex align-items-center">
                    <Checkbox
                        id="prompt-draft"
                        name="draft"
                        checked={value.draft}
                        onChange={createInputChangeHandler('draft')}
                        label="Draft"
                    />
                    <small className="text-muted">
                        &nbsp;&mdash; marking as draft means other people shouldn't use it yet
                    </small>
                </div>
                {afterFields}
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

const PromptTextField: FunctionComponent<
    Pick<ComponentProps<typeof TextArea>, 'label' | 'name' | 'value' | 'onChange'> & {
        richPromptEditor: boolean
    }
> = ({ label, name, value, onChange, richPromptEditor }) => {
    const promptEditorConfig = useMemo<PromptEditorConfig>(
        () => ({
            commandComponents: {
                Command,
                CommandEmpty,
                CommandGroup,
                CommandInput,
                CommandItem,
                CommandList,
                CommandLoading,
                CommandSeparator,
            },
        }),
        []
    )
    const extensionAPI = useMemo<ExtensionAPI
    return richPromptEditor ? (
        <ClientStateContextProvider value={{ initialContext: [] }}>
            <ExtensionAPIProviderForTestsOnly value={}
            <PromptEditorConfigProvider value={promptEditorConfig}>
                <PromptEditor />
            </PromptEditorConfigProvider>
        </ClientStateContextProvider>
    ) : (
        <TextArea label={label} name={name} value={value} onChange={onChange} rows={10} resizeable={true} />
    )
}
