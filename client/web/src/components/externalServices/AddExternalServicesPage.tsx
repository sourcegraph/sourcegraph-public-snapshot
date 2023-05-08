import { FC } from 'react'

import { mdiInformation } from '@mdi/js'
import { useLocation } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useLocalStorage, Button, Link, Alert, H2, H3, Icon, Text, Container } from '@sourcegraph/wildcard'

import { PageTitle } from '../PageTitle'

import { AddExternalServicePage } from './AddExternalServicePage'
import { ExternalServiceCard } from './ExternalServiceCard'
import { allExternalServices, AddExternalServiceOptions } from './externalServices'

import styles from './AddExternalServicesPage.module.scss'

export interface AddExternalServicesPageProps extends TelemetryProps {
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

    externalServicesFromFile: boolean
    allowEditExternalServicesWithFile: boolean

    /** For testing only. */
    autoFocusForm?: boolean
}

/**
 * Page for choosing a service kind and variant to add, among the available options.
 */
export const AddExternalServicesPage: FC<AddExternalServicesPageProps> = ({
    codeHostExternalServices,
    nonCodeHostExternalServices,
    telemetryService,
    autoFocusForm,
    externalServicesFromFile,
    allowEditExternalServicesWithFile,
}) => {
    const { search } = useLocation()
    const [hasDismissedPrivacyWarning, setHasDismissedPrivacyWarning] = useLocalStorage(
        'hasDismissedCodeHostPrivacyWarning',
        false
    )
    const dismissPrivacyWarning = (): void => {
        setHasDismissedPrivacyWarning(true)
    }

    const id = new URLSearchParams(search).get('id')
    if (id) {
        const externalService = allExternalServices[id]
        if (externalService) {
            return (
                <AddExternalServicePage
                    telemetryService={telemetryService}
                    externalService={externalService}
                    autoFocusForm={autoFocusForm}
                    externalServicesFromFile={externalServicesFromFile}
                    allowEditExternalServicesWithFile={allowEditExternalServicesWithFile}
                />
            )
        }
    }

    const licenseInfo = window.context.licenseInfo
    let allowedCodeHosts: AddExternalServiceOptions[] | null = null
    if (licenseInfo && licenseInfo.currentPlan === 'business-0') {
        allowedCodeHosts = [
            codeHostExternalServices.github,
            codeHostExternalServices.gitlabcom,
            codeHostExternalServices.bitbucket,
        ]
    }

    return (
        <>
            <PageTitle title="Add code host connection" />
            <H2>Add code host connection</H2>
            <Container>
                <Text>Add code host connection to one of the supported code hosts.</Text>
                {hasDismissedPrivacyWarning && (
                    <Alert variant="info">
                        <Text>
                            This Sourcegraph installation will never send your code, repository names, file names, or
                            any other specific code data to Sourcegraph.com or any other destination. Your code is kept
                            private on this installation.
                        </Text>
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
                                <Link to="/help/admin/permissions" target="_blank" rel="noopener noreferrer">
                                    user repository access permissions
                                </Link>
                                , if you have enabled this feature.
                            </li>
                            <li>
                                Opening pull requests and syncing their metadata as part of{' '}
                                <Link to="/help/batch_changes" target="_blank" rel="noopener noreferrer">
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
                    .filter(externalService => !allowedCodeHosts || allowedCodeHosts.includes(externalService[1]))
                    .sort(([, externalService1], [, externalService2]) =>
                        externalService1.title.localeCompare(externalService2.title)
                    )
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
                                externalService => allowedCodeHosts && !allowedCodeHosts.includes(externalService[1])
                            )
                            .sort(([, externalService1], [, externalService2]) =>
                                externalService1.title.localeCompare(externalService2.title)
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
            </Container>
        </>
    )
}

function getAddURL(id: string): string {
    if (id === 'ghapp') {
        return '../../github-apps/new'
    }

    const parameters = new URLSearchParams()
    parameters.append('id', id)
    return `?${parameters.toString()}`
}
