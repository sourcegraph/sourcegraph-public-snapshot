import * as React from 'react'
import { ActivationChecklist } from '../../../shared/src/components/activation/ActivationChecklist'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import H from 'history'

interface Props extends ActivationProps {
    history: H.History
    isLightTheme: boolean
}
interface State {}

export class OnboardingPage extends React.Component<Props, State> {
    public render(): JSX.Element | null {
        return (
            <div className="welcome-page-left">
                <div className="welcome-page-left__content-2">
                    <h1>Get started</h1>
                    <p>Complete these remaining steps to finish onboarding to Sourcegraph:</p>
                    {this.props.activation && (
                        <ActivationChecklist
                            history={this.props.history}
                            steps={this.props.activation.steps}
                            completed={this.props.activation.completed}
                        />
                    )}
                </div>
            </div>
        )
    }
}
