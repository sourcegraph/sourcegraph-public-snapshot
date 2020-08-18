import * as H from 'history'
import React from 'react'
import { PageTitle } from '../../../components/PageTitle'
import { ThemeProps } from '../../../../../shared/src/theme'
import { ExternalServiceCard } from '../../../components/ExternalServiceCard'
import {
    codeHostExternalServices,
    nonCodeHostExternalServices,
    allExternalServices,
} from '../../../site-admin/externalServices'
import { AddExternalServicePage } from './AddExternalServicePage'
import { useLocalStorage } from '../../../util/useLocalStorage'

interface Props extends ThemeProps {
    history: H.History
    eventLogger: {
        logViewEvent: (event: 'AddExternalService') => void
        log: (event: 'AddExternalServiceFailed' | 'AddExternalServiceSucceeded', eventProperties?: any) => void
    }
}

/**
 * Page for choosing a service kind and variant to add, among the available options.
 */
export const AddExternalServicesPage: React.FunctionComponent<Props> = props => {
    const [hasDismissedPrivacyWarning, setHasDismissedPrivacyWarning] = useLocalStorage(
        'hasDismissedCodeHostPrivacyWarning',
        false
    )
    const dismissPrivacyWarning = (): void => {
        setHasDismissedPrivacyWarning(true)
    }
    const id = new URLSearchParams(props.history.location.search).get('id')
    if (id) {
        const externalService = allExternalServices[id]
        if (externalService) {
            return <AddExternalServicePage {...props} externalService={externalService} />
        }
    }

    return (
        <div className="add-external-services-page mt-3">
            <PageTitle title="Add repositories" />
            <div className="d-flex justify-content-between align-items-center mt-3 mb-3">
                <h2 className="mb-0">Add repositories</h2>
            </div>
            <p className="mt-2">Add repositories from one of these code hosts.</p>
            {!hasDismissedPrivacyWarning && (
                <div className="alert alert-info">
                    <h3>This Sourcegraph installation will access your code host by:</h3>
                    <ul>
                        <li>
                            Periodically fetching a list of repositories to ensure new, removed, and renamed
                            repositories are accessible on Sourcegraph.
                        </li>
                        <li>Cloning the repositories you specify to create a local cache.</li>
                        <li>Periodically pulling cloned repositories to ensure search results are current.</li>
                        <li>
                            Fetching{' '}
                            <a
                                href="https://docs.sourcegraph.com/admin/repo/permissions"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                user repository access permissions
                            </a>
                            , if you have enabled this feature.
                        </li>
                        <li>
                            Opening pull requests and syncing their metadata as part of{' '}
                            <a
                                href="https://docs.sourcegraph.com/user/campaigns"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                code change campaigns
                            </a>
                            , if you have enabled this feature.
                        </li>
                    </ul>
                    <div className="d-flex justify-content-end">
                        <button className="btn btn-light" onClick={dismissPrivacyWarning} type="button">
                            Do not show this again
                        </button>
                    </div>
                </div>
            )}
            {Object.entries(codeHostExternalServices).map(([id, externalService]) => (
                <div className="add-external-services-page__card" key={id}>
                    <ExternalServiceCard to={getAddURL(id)} {...externalService} />
                </div>
            ))}
            <br />
            <h2>Other connections</h2>
            <p className="mt-2">Add connections to non-code-host services.</p>
            {Object.entries(nonCodeHostExternalServices).map(([id, externalService]) => (
                <div className="add-external-services-page__card" key={id}>
                    <ExternalServiceCard to={getAddURL(id)} {...externalService} />
                </div>
            ))}
        </div>
    )
}

function getAddURL(id: string): string {
    const parameters = new URLSearchParams()
    parameters.append('id', id)
    return `?${parameters.toString()}`
}
