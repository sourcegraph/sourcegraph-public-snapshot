import * as H from 'history'
import React, { FunctionComponent, useCallback, useEffect, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { ThemeProps } from '../../../../../shared/src/theme'
import { ErrorAlert } from '../../../components/alerts'
import { PageTitle } from '../../../components/PageTitle'
import { SettingsAreaRepositoryFields } from '../../../graphql-operations'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../settings/DynamicallyImportedMonacoSettingsEditor'
import { getConfiguration as defaultGetConfiguration, updateConfiguration } from './backend'
import allConfigSchema from './schema.json'

export interface CodeIntelIndexConfigurationPageProps extends RouteComponentProps<{}>, ThemeProps, TelemetryProps {
    repo: Pick<SettingsAreaRepositoryFields, 'id'>
    history: H.History
    getConfiguration?: typeof defaultGetConfiguration
}

export const CodeIntelIndexConfigurationPage: FunctionComponent<CodeIntelIndexConfigurationPageProps> = ({
    repo,
    isLightTheme,
    telemetryService,
    history,
    getConfiguration = defaultGetConfiguration,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelIndexConfigurationPage'), [telemetryService])

    const [fetchError, setFetchError] = useState<Error>()
    const [saveError, setSaveError] = useState<Error>()
    const [saving, setSaving] = useState(() => false)
    const [configuration, setConfiguration] = useState<string>()

    useEffect(() => {
        const subscription = getConfiguration({ id: repo.id }).subscribe(configuration => {
            setConfiguration(configuration?.indexConfiguration?.configuration || '')
        }, setFetchError)

        return () => subscription.unsubscribe()
    }, [repo, getConfiguration])

    const save = useCallback(
        async (content: string) => {
            setSaving(true)
            setSaveError(undefined)

            try {
                await updateConfiguration({ id: repo.id, content }).toPromise()
                setConfiguration(content)
            } catch (error) {
                setSaveError(error)
            } finally {
                setSaving(false)
            }
        },
        [repo]
    )

    return fetchError ? (
        <ErrorAlert prefix="Error fetching index configuration" error={fetchError} history={history} />
    ) : (
        <div className="code-intel-index-configuration web-content">
            <PageTitle title="Precise code intelligence index configuration" />
            <h2>Precise code intelligence index configuration</h2>
            <p>
                Override the inferred configuration when automatically indexing repositories on{' '}
                <a href="https://sourcegraph.com" target="_blank" rel="noreferrer noopener">
                    Sourcegraph.com
                </a>
                .
            </p>

            {saveError && <ErrorAlert prefix="Error saving index configuration" error={saveError} history={history} />}

            <DynamicallyImportedMonacoSettingsEditor
                value={configuration || ''}
                jsonSchema={allConfigSchema}
                canEdit={true}
                onSave={save}
                saving={saving}
                height={600}
                isLightTheme={isLightTheme}
                history={history}
                telemetryService={telemetryService}
            />
        </div>
    )
}
