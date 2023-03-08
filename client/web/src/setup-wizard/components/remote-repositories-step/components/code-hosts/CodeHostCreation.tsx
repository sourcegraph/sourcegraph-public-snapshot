import { FC, ReactNode, useMemo } from 'react'

import { useNavigate, useParams } from 'react-router-dom'

import { useMutation } from '@sourcegraph/http-client'
import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { Alert, Button, FormChangeEvent, H4, Link, useLocalStorage } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../../../../../components/externalServices/externalServices'
import { LoaderButton } from '../../../../../components/LoaderButton'
import {
    AddRemoteCodeHostResult,
    AddRemoteCodeHostVariables,
    GetCodeHostsResult,
} from '../../../../../graphql-operations'
import { getCodeHostKindFromURLParam } from '../../helpers'
import { ADD_CODE_HOST, CODE_HOST_FRAGMENT } from '../../queries'

import { CodeHostConnectFormFields, CodeHostJSONForm, CodeHostJSONFormState } from './common'
import { GithubConnectView } from './github/GithubConnectView'

import styles from './CodeHostCreation.module.scss'

/**
 * Renders creation UI for any supported code hosts (Github, Gitlab) based on
 * "codeHostType" URL param see root component routing logic.
 */
export const CodeHostCreation: FC = () => {
    const { codeHostType } = useParams()
    const codeHostKind = getCodeHostKindFromURLParam(codeHostType!)

    if (codeHostKind === null) {
        return (
            <Alert variant="warning">
                <H4>We either couldn't find "{codeHostType}" code host option or we do not support this</H4>
                Pick one of supported code host option <Link to="/setup/remote-repositories">here</Link>
            </Alert>
        )
    }

    // We render content inside react fragment because this view is rendered
    // within Container UI (avoid unnecessary DOM nesting)
    return (
        <CodeHostCreationView codeHostKind={codeHostKind}>
            {state => (
                <footer className={styles.footer}>
                    <LoaderButton
                        type="submit"
                        variant="primary"
                        size="sm"
                        label={state.submitting ? 'Connecting' : 'Connect'}
                        alwaysShowLabel={true}
                        loading={state.submitting}
                        disabled={state.submitting}
                    />
                    <Button as={Link} size="sm" to="/setup/remote-repositories" variant="secondary">
                        Cancel
                    </Button>
                </footer>
            )}
        </CodeHostCreationView>
    )
}

interface CodeHostCreationFormProps {
    codeHostKind: ExternalServiceKind
    children: (state: CodeHostJSONFormState) => ReactNode
}

/**
 * Renders specific creation UI form for particular code host type. Most of
 * the code hosts have similar UI but some of them (like GitHub) have special enhanced
 * UI with pickers and other form UI.
 */
const CodeHostCreationView: FC<CodeHostCreationFormProps> = props => {
    const { codeHostKind, children } = props

    const navigate = useNavigate()
    const externalServiceOptions = defaultExternalServices[codeHostKind]

    const defaultConnectionValues = useMemo(
        () => ({
            displayName: externalServiceOptions.defaultDisplayName,
            config: getDefaultConfiguration(externalServiceOptions.defaultConfig, externalServiceOptions.kind),
        }),
        [externalServiceOptions.defaultConfig, externalServiceOptions.defaultDisplayName, externalServiceOptions.kind]
    )

    const [localValues, setLocalValues] = useLocalStorage<CodeHostConnectFormFields>(
        `${codeHostKind}-connect-form`,
        defaultConnectionValues
    )

    const [addRemoteCodeHost] = useMutation<AddRemoteCodeHostResult, AddRemoteCodeHostVariables>(ADD_CODE_HOST)

    const handleFormChange = (event: FormChangeEvent<CodeHostConnectFormFields>): void => {
        if (event.valid) {
            setLocalValues(event.values)
        }
    }

    const handleFormSubmit = async (values: CodeHostConnectFormFields): Promise<void> => {
        await addRemoteCodeHost({
            variables: { input: { ...values, kind: codeHostKind } },
            refetchQueries: ['RepositoryStats', 'StatusMessages'],
            update: (cache, result) => {
                const { data } = result

                if (!data) {
                    return
                }

                cache.modify({
                    fields: {
                        externalServices(codeHosts: GetCodeHostsResult['externalServices']) {
                            const newCodeHost = cache.writeFragment({
                                data: data.addExternalService,
                                fragment: CODE_HOST_FRAGMENT,
                            })

                            // Update local cache and put newly created/connected code host
                            // to the beginning of code hosts list
                            return { nodes: [newCodeHost, ...(codeHosts?.nodes ?? [])] }
                        },
                    },
                })
            },
        })

        // Reset local storage values
        setLocalValues(defaultConnectionValues)
        navigate('/setup/remote-repositories')
        // TODO show notification UI that code host has been added successfully
    }

    if (codeHostKind === ExternalServiceKind.GITHUB) {
        return (
            <GithubConnectView initialValues={localValues} onChange={handleFormChange} onSubmit={handleFormSubmit}>
                {children}
            </GithubConnectView>
        )
    }

    return (
        <CodeHostJSONForm
            initialValues={localValues}
            externalServiceOptions={defaultExternalServices[codeHostKind]}
            onChange={handleFormChange}
            onSubmit={handleFormSubmit}
        >
            {children}
        </CodeHostJSONForm>
    )
}

const DEFAULT_GITHUB_CONNECTION_CONFIG = `
{
    "url": "https://github.com",
    "token": ""
}
`.trim()

function getDefaultConfiguration(defaultConfig: string, kind: ExternalServiceKind): string {
    // GitHub's connection form should have no orgs field by default
    // because it opens orgs option by default which should be hided in
    // setup wizard until user clicks the checkbox for organizations.
    if (kind === ExternalServiceKind.GITHUB) {
        return DEFAULT_GITHUB_CONNECTION_CONFIG
    }

    return defaultConfig
}
