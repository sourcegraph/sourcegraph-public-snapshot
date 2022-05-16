import React from 'react'

import * as H from 'history'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useLocalStorage, Button, Link, Alert, Typography } from '@sourcegraph/wildcard'

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

    return (
        <div className="add-external-services-page mt-3">
            <PageTitle title="Add repositories" />
            <div className="d-flex justify-content-between align-items-center mt-3 mb-3">
                <Typography.H2 className="mb-0">Add repositories</Typography.H2>
            </div>
            <p className="mt-2">Add repositories from one of these code hosts.</p>
            {!hasDismissedPrivacyWarning && (
                <Alert variant="info">
                    {!userID && (
                        <p>
                            This Sourcegraph installation will never send your code, repository names, file names, or
                            any other specific code data to Sourcegraph.com or any other destination. Your code is kept
                            private on this installation.
                        </p>
                    )}
                    <Typography.H3>This Sourcegraph installation will access your code host by:</Typography.H3>
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
            {Object.entries(codeHostExternalServices).map(([id, externalService]) => (
                <div className={styles.addExternalServicesPageCard} key={id}>
                    <ExternalServiceCard to={getAddURL(id)} {...externalService} />
                </div>
            ))}
            {Object.entries(nonCodeHostExternalServices).length > 0 && (
                <>
                    <br />
                    <Typography.H2>Other connections</Typography.H2>
                    <p className="mt-2">Add connections to non-code-host services.</p>
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
