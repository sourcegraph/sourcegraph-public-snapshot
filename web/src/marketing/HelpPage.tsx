import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../shared/src/graphql/schema'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { ThemeProps } from '../../../shared/src/theme'
import HelpCircleIcon from 'mdi-react/HelpCircleIcon'

interface SurveyPageProps extends RouteComponentProps<{ score?: string }>, ThemeProps {
    authenticatedUser: GQL.IUser | null
}

export class HelpPage extends React.Component<SurveyPageProps> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('Help')
    }

    public render(): JSX.Element | null {
        return (
            <div className="survey-page">
                <PageTitle title="Help" />
                <HeroPage
                    icon={HelpCircleIcon}
                    title="Need Help?"
                    body={
                        <div>
                            <div>
                                <p>Reach out to support if you have a question.</p>
                            </div>
                            <div>
                                <p>
                                    Take a look at our <a href="/documentation">documentation</a>.
                                </p>
                            </div>
                            <div>
                                <p>Send us your product feedback, ideas, and feature requests.</p>
                            </div>
                        </div>
                    }
                />
            </div>
        )
    }
}
