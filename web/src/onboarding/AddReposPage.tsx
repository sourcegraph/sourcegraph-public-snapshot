import * as H from 'history'
import * as React from 'react'
import { ExternalServiceCard } from '../components/ExternalServiceCard'
import { onboardingExternalServices } from '../site-admin/externalServices'
import { map } from 'lodash'
import { WelcomeAddExternalServicePage } from './AddExternalServicePage'

interface WelcomeAddReposPageProps {
    history: H.History
    eventLogger: {
        logViewEvent: (event: 'AddExternalService') => void
        log: (event: 'AddExternalServiceFailed' | 'AddExternalServiceSucceeded', eventProperties?: any) => void
    }
}
interface WelcomeAddReposPageState {}

export class WelcomeAddReposPage extends React.Component<WelcomeAddReposPageProps, WelcomeAddReposPageState> {
    public state: WelcomeAddReposPageState = {}

    public render(): JSX.Element | null {
        const externalServices = onboardingExternalServices

        const id = new URLSearchParams(this.props.history.location.search).get('id')
        if (id) {
            const externalService = externalServices[id]
            if (externalService) {
                return (
                    <WelcomeAddExternalServicePage
                        {...this.props}
                        externalService={externalService}
                        isLightTheme={true}
                    />
                )
            }
        }

        return (
            <div className="welcome-page-left">
                <div className="welcome-page-left__content">
                    <h2 className="welcome-page-left__content-header">
                        Where are the repositories you&rsquo;d like Sourcegraph to index?
                    </h2>
                    <p>
                        Note: Sourcegraph <b>never</b> sends your code to external servers.
                    </p>
                    {map(externalServices, (externalService, id) => (
                        <div className="add-external-services-page__card" key={id}>
                            <ExternalServiceCard to={`?id=${encodeURIComponent(id)}`} {...externalService} />
                        </div>
                    ))}
                    <br />
                    <br />
                </div>
            </div>
        )
    }
}
