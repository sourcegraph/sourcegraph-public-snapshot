import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { JSONSchema6 } from 'json-schema'
import React, { useCallback, useEffect, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subscription, Unsubscribable } from 'rxjs'
import { delay, mergeMap, retryWhen, tap, timeout } from 'rxjs/operators'
import siteSchemaJSON from '../../../schema/site.schema.json'
import * as GQL from '../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { JSONSchemaInstanceEditor } from '../components/jsonSchemaInstanceEditor/JSONSchemaInstanceEditor'
import { PageTitle } from '../components/PageTitle'
import { parseJSON } from '../settings/configuration'
import { refreshSiteFlags } from '../site/backend'
import { eventLogger } from '../tracking/eventLogger'
import { fetchSite, reloadSite, updateSiteConfiguration } from './backend'
import { SiteAdminManagementConsolePassword } from './SiteAdminManagementConsolePassword'

interface Props extends RouteComponentProps<{}> {
    isLightTheme: boolean
}

const LOADING = 'loading' as const

const EXPECTED_RELOAD_WAIT = 7 * 1000 // 7 seconds

/**
 * A page displaying the site configuration.
 */
export const SiteAdminConfigurationPage: React.FunctionComponent<Props> = props => {
    useEffect(() => eventLogger.logViewEvent('SiteAdminConfiguration'), [])

    const [restartToApply, setRestartToApply] = useState(window.context.needServerRestart)

    const [siteConfig, setSiteConfig] = useState<
        typeof LOADING | Pick<GQL.ISiteConfiguration, 'id' | 'validationMessages'> | ErrorLike
    >(LOADING)
    // Reduce jitter by optimistically preserving just-updated site config value in the UI.
    const [effectiveContents, setEffectiveContents] = useState<
        typeof LOADING | Pick<GQL.ISiteConfiguration, 'effectiveContents'>
    >(LOADING)
    const refreshSiteConfig = useCallback(() => {
        // tslint:disable-next-line: no-floating-promises
        ;(async () => {
            try {
                const siteConfig = (await fetchSite().toPromise()).configuration
                setSiteConfig(siteConfig)
                setEffectiveContents(siteConfig)
            } catch (err) {
                setSiteConfig(asError(err))
            }
        })()
    }, [])
    useEffect(() => refreshSiteConfig(), [refreshSiteConfig])

    const [updateOp, setUpdateOp] = useState<null | typeof LOADING | ErrorLike>(null)
    const updateSiteConfig = useCallback(
        (newContents: string) => {
            eventLogger.log('SiteConfigurationSaved')
            const lastConfigurationID = siteConfig !== LOADING && !isErrorLike(siteConfig) ? siteConfig.id : 0
            setUpdateOp(LOADING)
            // tslint:disable-next-line: no-floating-promises
            ;(async () => {
                try {
                    const restartToApply = await updateSiteConfiguration(lastConfigurationID, newContents).toPromise()
                    setUpdateOp(null)
                    if (restartToApply) {
                        window.context.needServerRestart = restartToApply
                    } else {
                        // Refresh site flags so that global site alerts reflect the latest
                        // configuration.
                        await refreshSiteFlags()
                    }
                    setRestartToApply(restartToApply)
                    refreshSiteConfig()
                } catch (err) {
                    setUpdateOp(asError(err))
                }
            })()
        },
        [refreshSiteConfig, siteConfig]
    )

    const [, forceUpdate] = useState()
    const [reload, setReload] = useState<{ startedAt: number; subscription: Unsubscribable } | ErrorLike>()
    const reloadSite2 = useCallback(() => {
        eventLogger.log('SiteReloaded')

        if (reload && !isErrorLike(reload)) {
            // Cancel last reload-wait.
            reload.subscription.unsubscribe()
        }

        const subscription = new Subscription()
        subscription.add(
            reloadSite()
                .pipe(
                    delay(2000),
                    mergeMap(() =>
                        // wait for server to restart
                        fetchSite().pipe(
                            retryWhen(x =>
                                x.pipe(
                                    tap(() => forceUpdate({})),
                                    delay(500)
                                )
                            ),
                            timeout(10000)
                        )
                    ),
                    tap(() => refreshSiteConfig())
                )
                .subscribe(
                    () => {
                        setReload(undefined)
                        window.location.reload() // brute force way to reload view state
                    },
                    err => setReload(asError(err))
                )
        )
        setReload({ startedAt: Date.now(), subscription })
    }, [refreshSiteConfig, reload])
    useEffect(
        () => () => {
            // Stop waiting for the reload when the user navigates away from this page.
            if (reload && !isErrorLike(reload)) {
                reload.subscription.unsubscribe()
            }
        },
        [reload]
    )

    return (
        <div className="site-admin-configuration-page">
            <PageTitle title="Site configuration" />
            <div className="d-flex justify-content-between align-items-center mt-3 mb-1">
                <h2 className="mb-0">Site configuration</h2>
            </div>
            <p>
                View and edit the Sourcegraph site configuration. See{' '}
                <Link to="/help/admin/config/site_config">documentation</Link> for more information.
            </p>
            <div className="mb-3">
                <SiteAdminManagementConsolePassword />
            </div>
            <p>
                Authentication providers, the application URL, license key, and other critical configuration may be
                edited via the <a href="https://docs.sourcegraph.com/admin/management_console">management console</a>.
            </p>
            <div className="site-admin-configuration-page__alerts">
                {isErrorLike(updateOp) && (
                    <div className="alert alert-danger site-admin-configuration-page__alert">
                        Error updating site configuration: {updateOp.message}
                    </div>
                )}
                {restartToApply && (
                    <div className="alert alert-warning site-admin-configuration-page__alert site-admin-configuration-page__alert-flex">
                        Server restart is required for the configuration to take effect.
                        <button className="btn btn-primary btn-sm" onClick={reloadSite2}>
                            Restart server
                        </button>
                    </div>
                )}
                {siteConfig !== LOADING &&
                    !isErrorLike(siteConfig) &&
                    siteConfig.validationMessages &&
                    siteConfig.validationMessages.length > 0 && (
                        <div className="alert alert-danger site-admin-configuration-page__alert">
                            <p>The server reported issues in the last-saved config:</p>
                            <ul>
                                {siteConfig.validationMessages.map((e, i) => (
                                    <li key={i} className="site-admin-configuration-page__alert-item">
                                        {e}
                                    </li>
                                ))}
                            </ul>
                        </div>
                    )}
                {reload && !isErrorLike(reload) && (
                    <div className="alert alert-primary site-admin-configuration-page__alert">
                        <div>
                            <LoadingSpinner className="icon-inline" /> Waiting for site to reload...
                        </div>
                        {Date.now() - reload.startedAt > EXPECTED_RELOAD_WAIT && (
                            <p className="mt-2">
                                <small>
                                    It's taking longer than expected. Check the server logs for error messages.
                                </small>
                            </p>
                        )}
                    </div>
                )}
            </div>
            {effectiveContents === LOADING ? (
                <LoadingSpinner className="icon-inline" />
            ) : isErrorLike(siteConfig) ? (
                <div className="alert alert-danger">Error loading site configuration: {siteConfig.message}</div>
            ) : (
                <div>
                    <JSONSchemaInstanceEditor
                        value={effectiveContents.effectiveContents}
                        jsonSchema={siteSchemaJSON}
                        canEdit={true}
                        saving={updateOp === LOADING}
                        loading={siteConfig === LOADING || (reload && !isErrorLike(reload)) || updateOp === LOADING}
                        height={600}
                        isLightTheme={props.isLightTheme}
                        onSave={updateSiteConfig}
                        history={props.history}
                        form={{
                            schema: siteSchemaJSON as JSONSchema6,
                            formData: parseJSON(effectiveContents.effectiveContents),
                        }}
                    />
                    <p className="form-text text-muted">
                        <small>
                            Use Ctrl+Space for completion, and hover over JSON properties for documentation. For more
                            information, see the <Link to="/help/admin/config/site_config">documentation</Link>.
                        </small>
                    </p>
                </div>
            )}
        </div>
    )
}
