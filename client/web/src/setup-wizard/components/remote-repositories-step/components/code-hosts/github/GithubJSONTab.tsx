import { FC, PropsWithChildren, ReactElement, useState } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Code, Collapse, CollapseHeader, CollapsePanel, Icon, Input, Link, Text } from '@sourcegraph/wildcard'

import { codeHostExternalServices } from '../../../../../../components/externalServices/externalServices'
import { FormGroup, getDefaultInputProps, useFieldAPI } from '../../../../../../enterprise/insights/components'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../../../../settings/DynamicallyImportedMonacoSettingsEditor'

import styles from './GithubJSONTab.module.scss'

interface GithubSettingsViewProps {
    displayNameField: useFieldAPI<string>
    configurationField: useFieldAPI<string>
}

export function GithubJSONTabView(props: GithubSettingsViewProps): ReactElement {
    const { displayNameField, configurationField } = props

    // Fragment to avoid nesting since it's rendered within TabPanel fieldset
    return (
        <>
            <Input label="Display name" placeholder="Github (Personal)" {...getDefaultInputProps(displayNameField)} />

            <FormGroup
                name="Configuration"
                title="Configuration"
                subtitle={<GithubInstructions />}
                labelClassName={styles.configurationGroupLabel}
            >
                <DynamicallyImportedMonacoSettingsEditor
                    // DynamicallyImportedMonacoSettingsEditor does not re-render the passed input.config
                    // if it thinks the config is dirty. We want to always replace the config if the kind changes
                    // so the editor is keyed on the kind.
                    value={configurationField.input.value}
                    actions={codeHostExternalServices.github.editorActions}
                    jsonSchema={codeHostExternalServices.github.jsonSchema}
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

const GithubInstructions: FC = () => {
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
            <CollapsePanel as="ol" className={styles.configurationGroupInstructions}>
                <li>
                    Set <Field>url</Field> to the URL of GitHub Enterprise.
                </li>
                <li>
                    Create a GitHub access token (
                    <Link
                        to="https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        instructions
                    </Link>
                    ) with <b>repo</b> scope.
                    <li>
                        Set the value of the <Field>token</Field> field as your access token, in the configuration
                        below.
                    </li>
                </li>
                <li>
                    Specify which repositories Sourcegraph should index using one of the following fields:
                    <ul>
                        <li>
                            <Field>orgs</Field>: a list of GitHub organizations.
                        </li>
                        <li>
                            <Field>repositoryQuery</Field>: a list of GitHub search queries.
                            <br />
                            For example,
                            <Value>"org:sourcegraph created:&gt;2019-11-01"</Value> selects all repositories in
                            organization "sourcegraph" created after November 1, 2019.
                            <br />
                            You may also use <Value>"affiliated"</Value> to select all repositories affiliated with the
                            access token.
                        </li>
                        <li>
                            <Field>repos</Field>: a list of individual repositories.
                        </li>
                    </ul>
                </li>
            </CollapsePanel>
        </Collapse>
    )
}

const Field: FC<PropsWithChildren<any>> = props => <Code className="hljs-type">{props.children}</Code>

const Value: FC<PropsWithChildren<any>> = props => <Code className="hljs-attr">{props.children}</Code>
