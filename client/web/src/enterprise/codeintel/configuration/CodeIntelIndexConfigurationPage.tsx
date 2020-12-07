import * as H from 'history'
import React, { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { ThemeProps } from '../../../../../shared/src/theme'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { PageTitle } from '../../../components/PageTitle'
import { SettingsAreaRepositoryFields } from '../../../graphql-operations'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../settings/DynamicallyImportedMonacoSettingsEditor'
import { getConfiguration, updateConfiguration } from './backend'
import allConfigSchema from './schema.json'

export interface CodeIntelIndexConfigurationPageProps extends RouteComponentProps<{}>, ThemeProps, TelemetryProps {
    repo: SettingsAreaRepositoryFields
    history: H.History
}

export const CodeIntelIndexConfigurationPage: FunctionComponent<CodeIntelIndexConfigurationPageProps> = ({
    repo,
    isLightTheme,
    telemetryService,
    history,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelIndexConfigurationPage'), [telemetryService])

    const [saving, setSaving] = useState(() => false)

    const configuration = useObservable(useMemo(() => getConfiguration({ id: repo.id }), [repo]))

    const save = useCallback(
        async (content: string) => {
            setSaving(true)
            await updateConfiguration({ id: repo.id, content }).toPromise()
            setSaving(false)
        },
        [repo]
    )

    return (
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

            <DynamicallyImportedMonacoSettingsEditor
                value={configuration?.indexConfiguration?.configuration || ''}
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
