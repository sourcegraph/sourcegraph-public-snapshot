import { FC, ReactElement, ReactNode, useState } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Button,
    Collapse,
    CollapseHeader,
    CollapsePanel,
    Icon,
    Input,
    Text,
    useLocalStorage,
} from '@sourcegraph/wildcard'

import { AddExternalServiceOptions } from '../../../../../../components/externalServices/externalServices'
import {
    FormGroup,
    getDefaultInputProps,
    SubmissionErrors,
    useField,
    useFieldAPI,
    useForm,
} from '../../../../../../enterprise/insights/components'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../../../../settings/DynamicallyImportedMonacoSettingsEditor'

import styles from './CodeHostConnection.module.scss'

export interface CodeHostConnectFormFields {
    displayName: string
    configuration: string
}

export interface CodeHostJSONFormState {
    submitting: boolean
    submitErrors: SubmissionErrors
}

interface CodeHostJSONFormProps {
    externalServiceOptions: AddExternalServiceOptions
    initialValues?: CodeHostConnectFormFields
    children: (state: CodeHostJSONFormState) => ReactNode
    onSubmit: (values: CodeHostConnectFormFields) => Promise<void>
}

export function CodeHostJSONForm(props: CodeHostJSONFormProps): ReactElement {
    const { externalServiceOptions, initialValues, onSubmit, children } = props

    const [localValues, setLocalValues] = useLocalStorage<CodeHostConnectFormFields>(
        `${externalServiceOptions.kind}-connect-form`,
        {
            displayName: externalServiceOptions.defaultDisplayName,
            configuration: externalServiceOptions.defaultConfig,
        }
    )

    const form = useForm<CodeHostConnectFormFields>({
        initialValues: initialValues ?? localValues,
        onSubmit,
        onChange: event => setLocalValues(event.values),
    })

    const displayName = useField({
        formApi: form.formAPI,
        name: 'displayName',
        required: true,
    })

    const configuration = useField({
        formApi: form.formAPI,
        name: 'configuration',
    })

    return (
        // eslint-disable-next-line react/forbid-elements
        <form ref={form.ref} className={styles.formView} onSubmit={form.handleSubmit}>
            <CodeHostJSONFormContent
                displayNameField={displayName}
                configurationField={configuration}
                externalServiceOptions={externalServiceOptions}
            />

            <div className={styles.footer}>{children(form.formAPI)}</div>
        </form>
    )
}

interface CodeHostJSONFormContentProps {
    displayNameField: useFieldAPI<string>
    configurationField: useFieldAPI<string>
    externalServiceOptions: AddExternalServiceOptions
}

export function CodeHostJSONFormContent(props: CodeHostJSONFormContentProps): ReactElement {
    const { displayNameField, configurationField, externalServiceOptions } = props

    // Fragment to avoid nesting since it's rendered within TabPanel fieldset
    return (
        <>
            <Input label="Display name" {...getDefaultInputProps(displayNameField)} />

            <FormGroup
                name="Configuration"
                title="Configuration"
                subtitle={<CodeHostInstructions instructions={externalServiceOptions.instructions} />}
                labelClassName={styles.configurationGroupLabel}
            >
                <DynamicallyImportedMonacoSettingsEditor
                    // DynamicallyImportedMonacoSettingsEditor does not re-render the passed input.config
                    // if it thinks the config is dirty. We want to always replace the config if the kind changes
                    // so the editor is keyed on the kind.
                    value={configurationField.input.value}
                    actions={externalServiceOptions.editorActions}
                    jsonSchema={externalServiceOptions.jsonSchema}
                    canEdit={false}
                    loading={true}
                    height={400}
                    readOnly={false}
                    isLightTheme={true}
                    onChange={configurationField.input.onChange}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    className={styles.configurationGroupEditor}
                    explanation={
                        <Text className="form-text text-muted" size="small">
                            Use Ctrl+Space for completion, and hover over JSON properties for documentation.
                        </Text>
                    }
                />
            </FormGroup>
        </>
    )
}

interface CodeHostInstructionsProps {
    instructions: ReactNode
}

const CodeHostInstructions: FC<CodeHostInstructionsProps> = props => {
    const { instructions } = props
    const [isInstructionOpen, setInstructionOpen] = useState(false)

    return (
        <Collapse isOpen={isInstructionOpen} onOpenChange={setInstructionOpen}>
            <CollapseHeader
                as={Button}
                outline={false}
                variant="link"
                size="sm"
                className={styles.configurationGroupInstructionButton}
            >
                See instructions how to fill out JSONC configuration{' '}
                <Icon aria-hidden={true} svgPath={isInstructionOpen ? mdiChevronDown : mdiChevronUp} className="mr-1" />
            </CollapseHeader>
            <CollapsePanel className={styles.configurationGroupInstructions}>{instructions}</CollapsePanel>
        </Collapse>
    )
}
