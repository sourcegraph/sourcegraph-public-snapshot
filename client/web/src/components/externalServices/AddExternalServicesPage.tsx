import React from 'react'

import { mdiInformation } from '@mdi/js'
import * as H from 'history'

import { ExternalServiceKind } from '@sourcegraph/shared/src/schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useLocalStorage, Button, Link, Alert, H2, H3, Icon, Text } from '@sourcegraph/wildcard'

import { Scalars } from '../../graphql-operations'
import { PageTitle } from '../PageTitle'

import { AddExternalServicePage } from './AddExternalServicePage'
import { ExternalServiceCard } from './ExternalServiceCard'
import { allExternalServices, AddExternalServiceOptions } from './externalServices'

import styles from './AddExternalServicesPage.module.scss'

export interface AddExternalServicesPageProps extends ThemeProps, TelemetryProps {
    history: H.History
    routingPrefix: string
    afterCreateRoute: string
    userID?: Scalars['ID']

    /**
     * The list of code host external services to be displayed.
     * Pick items from externalServices.codeHostExternalServices.
     */
    codeHostExternalServices: Record<string, AddExternalServiceOptions>
    /**
     * The list of non-code host external services to be displayed.
     * Pick items from externalServices.nonCodeHostExternalServices.
     */
    nonCodeHostExternalServices: Record<string, AddExternalServiceOptions>

    /** For testing only. */
    autoFocusForm?: boolean
}

/**
 * Page for choosing a service kind and variant to add, among the available options.
 */
export const AddExternalServicesPage: React.FunctionComponent<
    React.PropsWithChildren<AddExternalServicesPageProps>
> = ({
    afterCreateRoute,
    codeHostExternalServices,
    history,
    isLightTheme,
    nonCodeHostExternalServices,
    routingPrefix,
    telemetryService,
    userID,
    autoFocusForm,
}) => {
    const [hasDismissedPrivacyWarning, setHasDismissedPrivacyWarning] = useLocalStorage(
        'hasDismissedCodeHostPrivacyWarning',
        false
    )
    const dismissPrivacyWarning = (): void => {
        setHasDismissedPrivacyWarning(true)
    }
    const id = new URLSearchParams(history.location.search).get('id')
    if (id) {
        const externalService = allExternalServices[id]
        if (externalService) {
            return (
                <AddExternalServicePage
                    afterCreateRoute={afterCreateRoute}
                    history={history}
                    isLightTheme={isLightTheme}
                    routingPrefix={routingPrefix}
                    telemetryService={telemetryService}
                    userID={userID}
                    externalService={externalService}
                    autoFocusForm={autoFocusForm}
                />
            )
        }
    }

    const licenseInfo = window.context.licenseInfo
    let allowedCodeHosts: ExternalServiceKind[] | null = null
    if (licenseInfo && licenseInfo.currentPlan === 'business-0') {
        allowedCodeHosts = [ExternalServiceKind.GITHUB, ExternalServiceKind.GITLAB, ExternalServiceKind.BITBUCKETCLOUD]
    }

    return (
        <div className="add-external-services-page mt-3">
            <PageTitle title="Add repositories" />
            <div className="d-flex justify-content-between align-items-center mt-3 mb-3">
                <H2 className="mb-0">Add repositories</H2>
            </div>
            <Text className="mt-2">Add repositories from one of these code hosts.</Text>
            {!hasDismissedPrivacyWarning && (
                <Alert variant="info">
                    {!userID && (
                        <Text>
                            This Sourcegraph installation will never send your code, repository names, file names, or
                            any other specific code data to Sourcegraph.com or any other destination. Your code is kept
                            private on this installation.
                        </Text>
                    )}
                    <H3>This Sourcegraph installation will access your code host by:</H3>
                    <ul>
                        <li>
                            Periodically fetching a list of repositories to ensure new, removed, and renamed
                            repositories are accessible on Sourcegraph.
                        </li>
                        <li>Cloning the repositories you specify to create a local cache.</li>
                        <li>Periodically pulling cloned repositories to ensure search results are current.</li>
                        <li>
                            Fetching{' '}
                            <Link
                                to="https://docs.sourcegraph.com/admin/repo/permissions"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                user repository access permissions
                            </Link>
                            , if you have enabled this feature.
                        </li>
                        <li>
                            Opening pull requests and syncing their metadata as part of{' '}
                            <Link
                                to="https://docs.sourcegraph.com/user/batch_changes"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                batch changes
                            </Link>
                            , if you have enabled this feature.
                        </li>
                    </ul>
                    <div className="d-flex justify-content-end">
                        <Button variant="secondary" className={styles.btnLight} onClick={dismissPrivacyWarning}>
                            Do not show this again
                        </Button>
                    </div>
                </Alert>
            )}
            {Object.entries(codeHostExternalServices)
                .filter(externalService => !allowedCodeHosts || allowedCodeHosts.includes(externalService[1].kind))
                .map(([id, externalService]) => (
                    <div className={styles.addExternalServicesPageCard} key={id}>
                        <ExternalServiceCard to={getAddURL(id)} {...externalService} />
                    </div>
                ))}
            {allowedCodeHosts && (
                <>
                    <br />
                    <Text>
                        <Icon aria-label="Information icon" svgPath={mdiInformation} /> Upgrade to{' '}
                        <Link to="https://about.sourcegraph.com/pricing">Sourcegraph Enterprise</Link> to add
                        repositories from other code hosts.
                    </Text>
                    {Object.entries(codeHostExternalServices)
                        .filter(
                            externalService => allowedCodeHosts && !allowedCodeHosts.includes(externalService[1].kind)
                        )
                        .map(([id, externalService]) => (
                            <div className={styles.addExternalServicesPageCard} key={id}>
                                <ExternalServiceCard
                                    to={getAddURL(id)}
                                    {...externalService}
                                    enabled={false}
                                    badge="enterprise"
                                    tooltip="Upgrade to Sourcegraph Enterprise to add repositories from other code hosts"
                                />
                            </div>
                        ))}
                </>
            )}
            {Object.entries(nonCodeHostExternalServices).length > 0 && (
                <>
                    <br />
                    <H2>Other connections</H2>
                    <Text className="mt-2">Add connections to non-code-host services.</Text>
                    {Object.entries(nonCodeHostExternalServices).map(([id, externalService]) => (
                        <div className={styles.addExternalServicesPageCard} key={id}>
                            <ExternalServiceCard to={getAddURL(id)} {...externalService} />
                        </div>
                    ))}
                </>
            )}
        </div>
    )
}

function getAddURL(id: string): string {
    const parameters = new URLSearchParams()
    parameters.append('id', id)
    return `?${parameters.toString()}`
}
