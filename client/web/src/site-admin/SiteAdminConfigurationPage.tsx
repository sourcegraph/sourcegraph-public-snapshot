import { type FC, useState, useEffect, useCallback } from 'react'

import { FetchResult, useApolloClient } from '@apollo/client'
import classNames from 'classnames'
import * as jsonc from 'jsonc-parser'
import { useSearchParams } from 'react-router-dom'

import { logger } from '@sourcegraph/common'
import { useQuery, useMutation } from '@sourcegraph/http-client'
import { type SiteConfiguration } from '@sourcegraph/shared/src/schema/site.schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import {
    Button,
    LoadingSpinner,
    Link,
    Alert,
    Code,
    Text,
    PageHeader,
    Container,
    ErrorAlert,
    Tab,
    TabList,
    TabPanel,
    TabPanels,
    Tabs,
} from '@sourcegraph/wildcard'

import siteSchemaJSON from '../../../../schema/site.schema.json'
import { AuthenticatedUser } from '../auth'
import { PageTitle } from '../components/PageTitle'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import {
    type ReloadSiteResult,
    type ReloadSiteVariables,
    type SiteResult,
    type SiteVariables,
    type UpdateSiteConfigurationResult,
    type UpdateSiteConfigurationVariables,
} from '../graphql-operations'
import { DynamicallyImportedMonacoSettingsEditor } from '../settings/DynamicallyImportedMonacoSettingsEditor'
import { refreshSiteFlags } from '../site/backend'
import { eventLogger } from '../tracking/eventLogger'

import { RELOAD_SITE, SITE_CONFIG_QUERY, UPDATE_SITE_CONFIG } from './backend'
import { SiteConfigurationChangeList } from './SiteConfigurationChangeList'
import { SMTPConfigForm } from './smtp/ConfigForm'

import styles from './SiteAdminConfigurationPage.module.scss'

export const defaultModificationOptions: jsonc.ModificationOptions = {
    formattingOptions: {
        eol: '\n',
        insertSpaces: true,
        tabSize: 2,
    },
}

function editWithComments(
    config: string,
    path: jsonc.JSONPath,
    value: any,
    comments: { [key: string]: string }
): jsonc.Edit {
    const edit = jsonc.modify(config, path, value, defaultModificationOptions)[0]
    for (const commentKey of Object.keys(comments)) {
        edit.content = edit.content.replace(`"${commentKey}": true,`, comments[commentKey])
        edit.content = edit.content.replace(`"${commentKey}": true`, comments[commentKey])
    }
    return edit
}

const quickConfigureActions: {
    id: string
    label: string
    run: (config: string) => { edits: jsonc.Edit[]; selectText: string }
}[] = [
    {
        id: 'setExternalURL',
        label: 'Set external URL',
        run: config => {
            const value = '<external URL>'
            const edits = jsonc.modify(config, ['externalURL'], value, defaultModificationOptions)
            return { edits, selectText: '<external URL>' }
        },
    },
    {
        id: 'setLicenseKey',
        label: 'Set license key',
        run: config => {
            const value = '<license key>'
            const edits = jsonc.modify(config, ['licenseKey'], value, defaultModificationOptions)
            return { edits, selectText: '<license key>' }
        },
    },
    {
        id: 'addGitLabAuth',
        label: 'Add GitLab sign-in',
        run: config => {
            const edits = [
                editWithComments(
                    config,
                    ['auth.providers', -1],
                    {
                        COMMENT: true,
                        type: 'gitlab',
                        displayName: 'GitLab',
                        url: '<GitLab URL>',
                        clientID: '<client ID>',
                        clientSecret: '<client secret>',
                    },
                    {
                        COMMENT: '// See https://docs.sourcegraph.com/admin/auth#gitlab for instructions',
                    }
                ),
            ]
            return { edits, selectText: '<GitLab URL>' }
        },
    },
    {
        id: 'addGitHubAuth',
        label: 'Add GitHub sign-in',
        run: config => {
            const edits = [
                editWithComments(
                    config,
                    ['auth.providers', -1],
                    {
                        COMMENT: true,
                        type: 'github',
                        displayName: 'GitHub',
                        url: 'https://github.com/',
                        allowSignup: true,
                        clientID: '<client ID>',
                        clientSecret: '<client secret>',
                    },
                    { COMMENT: '// See https://docs.sourcegraph.com/admin/auth#github for instructions' }
                ),
            ]
            return { edits, selectText: '<client ID>' }
        },
    },
    {
        id: 'useOneLoginSAML',
        label: 'Add OneLogin SAML',
        run: config => {
            const edits = [
                editWithComments(
                    config,
                    ['auth.providers', -1],
                    {
                        COMMENT: true,

                        type: 'saml',
                        displayName: 'OneLogin',
                        identityProviderMetadataURL: '<identity provider metadata URL>',
                    },
                    {
                        COMMENT: '// See https://docs.sourcegraph.com/admin/auth/saml/one_login for instructions',
                    }
                ),
            ]
            return { edits, selectText: '<identity provider metadata URL>' }
        },
    },
    {
        id: 'useOktaSAML',
        label: 'Add Okta SAML',
        run: config => {
            const value = {
                COMMENT: true,
                type: 'saml',
                displayName: 'Okta',
                identityProviderMetadataURL: '<identity provider metadata URL>',
            }
            const edits = [
                editWithComments(config, ['auth.providers', -1], value, {
                    COMMENT: '// See https://docs.sourcegraph.com/admin/auth/saml/okta for instructions',
                }),
            ]
            return { edits, selectText: '<identity provider metadata URL>' }
        },
    },
    {
        id: 'useSAML',
        label: 'Add other SAML',
        run: config => {
            const edits = [
                editWithComments(
                    config,
                    ['auth.providers', -1],
                    {
                        COMMENT: true,
                        type: 'saml',
                        displayName: 'SAML',
                        identityProviderMetadataURL: '<SAML IdP metadata URL>',
                    },
                    { COMMENT: '// See https://docs.sourcegraph.com/admin/auth/saml for instructions' }
                ),
            ]
            return { edits, selectText: '<SAML IdP metadata URL>' }
        },
    },
    {
        id: 'useOIDC',
        label: 'Add OpenID Connect',
        run: config => {
            const edits = [
                editWithComments(
                    config,
                    ['auth.providers', -1],
                    {
                        COMMENT: true,
                        type: 'openidconnect',
                        displayName: 'OpenID Connect',
                        issuer: '<identity provider URL>',
                        clientID: '<client ID>',
                        clientSecret: '<client secret>',
                    },
                    { COMMENT: '// See https://docs.sourcegraph.com/admin/auth#openid-connect for instructions' }
                ),
            ]
            return { edits, selectText: '<identity provider URL>' }
        },
    },
]

interface Props extends TelemetryProps {
    isSourcegraphApp: boolean
    authenticatedUser: AuthenticatedUser
}

const EXPECTED_RELOAD_WAIT = 7 * 1000 // 7 seconds
const SITE_CONFIG_QUERY_NAME = 'Site'

/**
 * A page displaying the site configuration.
 */
export const SiteAdminConfigurationPage: FC<Props> = ({ authenticatedUser, isSourcegraphApp, telemetryService }) => {
    const client = useApolloClient()
    const [params, setSearchParams] = useSearchParams()
    const [tabIndex, setTabIndex] = useState(Number(params.get('tab')) ?? 0)
    const [reloadStartedAt, setReloadStartedAt] = useState<Date>(new Date(0))
    const [enabledCompletions, setEnabledCompletions] = useState(false)
    const isLightTheme = useIsLightTheme()

    useEffect(() => eventLogger.logViewEvent('SiteAdminConfiguration'))

    const [isSetupChecklistEnabled] = useFeatureFlag('setup-checklist', false)

    useEffect(() => {
        if (isSetupChecklistEnabled && Number(params.get('tab')) !== tabIndex) {
            setSearchParams({ tab: tabIndex.toString() })
        }
    }, [tabIndex, isSetupChecklistEnabled, params, setSearchParams])

    const { data, loading, error } = useQuery<SiteResult, SiteVariables>(SITE_CONFIG_QUERY, {
        // fetchPolicy: 'cache-and-network',
    })

    const [reloadSiteConfig, { loading: reloadLoading, error: reloadError }] = useMutation<
        ReloadSiteResult,
        ReloadSiteVariables
    >(RELOAD_SITE)

    const [updateSiteConfig, { loading: updateLoading, error: updateError }] = useMutation<
        UpdateSiteConfigurationResult,
        UpdateSiteConfigurationVariables
    >(UPDATE_SITE_CONFIG, {
        refetchQueries: [SITE_CONFIG_QUERY_NAME],
    })

    const reloadSite = useCallback((): Promise<FetchResult<ReloadSiteResult>> => {
        eventLogger.log('SiteReloaded')
        setReloadStartedAt(new Date())
        return reloadSiteConfig()
    }, [setReloadStartedAt, reloadSiteConfig])

    const onSave = useCallback(
        async (newContents: string): Promise<void> => {
            eventLogger.log('SiteConfigurationSaved')

            const lastConfiguration = data?.site?.configuration
            const lastConfigurationID = lastConfiguration?.id || 0

            const result = await updateSiteConfig({
                variables: {
                    lastID: lastConfigurationID,
                    input: newContents,
                },
            })

            const restartToApply = result.data?.updateSiteConfiguration

            const oldContents = lastConfiguration?.effectiveContents || ''
            const oldConfiguration = jsonc.parse(oldContents) as SiteConfiguration
            const newConfiguration = jsonc.parse(newContents) as SiteConfiguration

            // Flipping these feature flags require a reload for the
            // UI to be rendered correctly in the navbar and the sidebar.
            const keys: (keyof SiteConfiguration)[] = ['batchChanges.enabled', 'codeIntelAutoIndexing.enabled']

            if (!keys.every(key => !!oldConfiguration?.[key] === !!newConfiguration?.[key])) {
                window.location.reload()
            }

            setEnabledCompletions(!oldConfiguration?.completions?.enabled && !!newConfiguration?.completions?.enabled)

            if (restartToApply) {
                window.context.needServerRestart = restartToApply
            } else {
                // Refresh site flags so that global site alerts
                // reflect the latest configuration.
                try {
                    await refreshSiteFlags(client)
                } catch (error) {
                    logger.error(error)
                }
            }
        },
        [client, data, setEnabledCompletions, updateSiteConfig]
    )

    let effectiveError: Error | undefined = error || reloadError
    if (updateError) {
        effectiveError =
            effectiveError ||
            new Error(
                String(updateError) +
                    '\nError occured while attempting to save site configuration. Please backup changes before reloading the page.'
            )
    }

    const alerts: JSX.Element[] = []
    if (effectiveError) {
        alerts.push(<ErrorAlert key="error" className={styles.alert} error={effectiveError} />)
    }
    if (reloadLoading) {
        alerts.push(
            <Alert key="error" className={styles.alert} variant="primary">
                <Text>
                    <LoadingSpinner /> Waiting for site to reload...
                </Text>
                {Date.now() - reloadStartedAt.valueOf() > EXPECTED_RELOAD_WAIT && (
                    <Text>
                        <small>It's taking longer than expected. Check the server logs for error messages.</small>
                    </Text>
                )}
            </Alert>
        )
    }
    if (window.context.needServerRestart) {
        alerts.push(
            <Alert key="remote-dirty" className={classNames(styles.alert, styles.alertFlex)} variant="warning">
                Server restart is required for the configuration to take effect.
                {(!data?.site || data.site?.canReloadSite) && (
                    <Button onClick={reloadSite} variant="primary" size="sm">
                        Restart server
                    </Button>
                )}
            </Alert>
        )
    }
    if (data?.site?.configuration?.validationMessages && data?.site.configuration.validationMessages.length > 0) {
        alerts.push(
            <Alert key="validation-messages" className={styles.alert} variant="danger">
                <Text>The server reported issues in the last-saved config:</Text>
                <ul>
                    {data?.site.configuration.validationMessages.map((message, index) => (
                        <li key={index} className={styles.alertItem}>
                            {message}
                        </li>
                    ))}
                </ul>
            </Alert>
        )
    }

    // Avoid user confusion with values.yaml properties mixed in with site config properties.
    const contents = data?.site?.configuration?.effectiveContents
    const legacyKubernetesConfigProps = [
        'alertmanagerConfig',
        'alertmanagerURL',
        'authProxyIP',
        'authProxyPassword',
        'deploymentOverrides',
        'gitoliteIP',
        'gitserverCount',
        'gitserverDiskSize',
        'gitserverSSH',
        'httpNodePort',
        'httpsNodePort',
        'indexedSearchDiskSize',
        'langGo',
        'langJava',
        'langJavaScript',
        'langPHP',
        'langPython',
        'langSwift',
        'langTypeScript',
        'nodeSSDPath',
        'phabricatorIP',
        'prometheus',
        'pyPIIP',
        'rbac',
        'storageClass',
        'useAlertManager',
    ].filter(property => contents?.includes(`"${property}"`))
    if (legacyKubernetesConfigProps.length > 0) {
        alerts.push(
            <Alert key="legacy-cluster-props-present" className={styles.alert} variant="info">
                The configuration contains properties that are valid only in the
                <Code>values.yaml</Code> config file used for Kubernetes cluster deployments of Sourcegraph:{' '}
                <Code>{legacyKubernetesConfigProps.join(' ')}</Code>. You can disregard the validation warnings for
                these properties reported by the configuration editor.
            </Alert>
        )
    }

    if (enabledCompletions) {
        alerts.push(
            <Alert key="cody-beta-notice" className={styles.alert} variant="info">
                By turning on completions for "Cody beta," you have read the{' '}
                <Link to="/help/cody">Cody Documentation</Link> and agree to the{' '}
                <Link to="https://about.sourcegraph.com/terms/cody-notice">Cody Notice and Usage Policy</Link>. In
                particular, some code snippets will be sent to a third-party language model provider when you use Cody
                questions.
            </Alert>
        )
    }

    return (
        <div>
            <PageTitle title="Configuration - Admin" />
            <PageHeader path={[{ text: 'Site configuration' }]} headingElement="h2" className="mb-3" />
            <div>{alerts}</div>
            {loading && <LoadingSpinner />}
            {isSetupChecklistEnabled && data && (
                <Tabs defaultIndex={tabIndex} onChange={setTabIndex} size="medium">
                    <TabList>
                        <Tab>Basic</Tab>
                        <Tab>JSON</Tab>
                    </TabList>
                    <TabPanels>
                        <TabPanel>
                            <Container className="mt-3">
                                <SMTPConfigForm
                                    authenticatedUser={authenticatedUser}
                                    config={data?.site?.configuration?.effectiveContents}
                                    saveConfig={onSave}
                                    loading={loading || updateLoading}
                                    error={updateError}
                                />
                            </Container>
                        </TabPanel>
                        <TabPanel>
                            {data?.site?.configuration && (
                                <Container className="mt-3">
                                    <Text className="mb-3">
                                        View and edit the Sourcegraph site configuration. See{' '}
                                        <Link target="_blank" to="/help/admin/config/site_config">
                                            documentation
                                        </Link>{' '}
                                        for more information.
                                    </Text>
                                    <DynamicallyImportedMonacoSettingsEditor
                                        value={contents || ''}
                                        jsonSchema={siteSchemaJSON}
                                        canEdit={true}
                                        saving={updateLoading}
                                        loading={loading || reloadLoading || updateLoading}
                                        height={600}
                                        isLightTheme={isLightTheme}
                                        onSave={onSave}
                                        actions={
                                            isSourcegraphApp || isSetupChecklistEnabled ? [] : quickConfigureActions
                                        }
                                        telemetryService={telemetryService}
                                        explanation={
                                            <Text className="form-text text-muted">
                                                <small>
                                                    Use Ctrl+Space for completion, and hover over JSON properties for
                                                    documentation. For more information, see the{' '}
                                                    <Link to="/help/admin/config/site_config">documentation</Link>.
                                                </small>
                                            </Text>
                                        }
                                    />
                                </Container>
                            )}
                        </TabPanel>
                    </TabPanels>
                </Tabs>
            )}
            {!isSetupChecklistEnabled && data?.site?.configuration && (
                <Container className="mt-3">
                    <Text className="mb-3">
                        View and edit the Sourcegraph site configuration. See{' '}
                        <Link target="_blank" to="/help/admin/config/site_config">
                            documentation
                        </Link>{' '}
                        for more information.
                    </Text>
                    <DynamicallyImportedMonacoSettingsEditor
                        value={contents || ''}
                        jsonSchema={siteSchemaJSON}
                        canEdit={true}
                        saving={updateLoading}
                        loading={loading || reloadLoading || updateLoading}
                        height={600}
                        isLightTheme={isLightTheme}
                        onSave={onSave}
                        actions={isSourcegraphApp || isSetupChecklistEnabled ? [] : quickConfigureActions}
                        telemetryService={telemetryService}
                        explanation={
                            <Text className="form-text text-muted">
                                <small>
                                    Use Ctrl+Space for completion, and hover over JSON properties for documentation. For
                                    more information, see the{' '}
                                    <Link to="/help/admin/config/site_config">documentation</Link>.
                                </small>
                            </Text>
                        }
                    />
                </Container>
            )}
            <div className="mt-3">
                <SiteConfigurationChangeList />
            </div>
        </div>
    )
}
