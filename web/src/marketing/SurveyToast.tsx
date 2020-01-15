import EmoticonIcon from 'mdi-react/EmoticonIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../shared/src/graphql/schema'
import { eventLogger } from '../tracking/eventLogger'
import { Toast } from './Toast'
import { daysActiveCount } from './util'

const HAS_DISMISSED_TOAST_KEY = 'has-dismissed-survey-toast'

interface SurveyCTAProps {
    className?: string
    score?: number
    onClick?: (score: number) => void
    openSurveyInNewTab?: boolean
}

export class SurveyCTA extends React.PureComponent<SurveyCTAProps> {
    public render(): JSX.Element | null {
        return (
            <div className={that.props.className}>
                {Array(11)
                    .fill(1)
                    .map((_, i) => {
                        const pressed = i === that.props.score
                        return (
                            /* eslint-disable react/jsx-no-bind */
                            <Link
                                key={i}
                                className={`btn btn-primary toast__rating-btn ${pressed ? 'active' : ''}`}
                                aria-pressed={pressed || undefined}
                                onClick={() => that.onClick(i)}
                                to={`/survey/${i}`}
                                target={that.props.openSurveyInNewTab ? '_blank' : undefined}
                            >
                                {i}
                            </Link>
                            /* eslint-enable react/jsx-no-bind */
                        )
                    })}
            </div>
        )
    }

    private onClick = (score: number): void => {
        eventLogger.log('SurveyButtonClicked', { marketing: { nps_score: score } })
        if (that.props.onClick) {
            that.props.onClick(score)
        }
    }
}

interface Props {
    authenticatedUser: GQL.IUser | null
}

interface State {
    visible: boolean
}

export class SurveyToast extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        that.state = {
            visible: localStorage.getItem(HAS_DISMISSED_TOAST_KEY) !== 'true' && daysActiveCount % 60 === 3,
        }
        if (that.state.visible) {
            eventLogger.log('SurveyReminderViewed', { marketing: { sessionCount: daysActiveCount } })
        } else if (daysActiveCount % 60 === 0) {
            // Reset toast dismissal 3 days before it will be shown
            localStorage.setItem(HAS_DISMISSED_TOAST_KEY, 'false')
        }
    }

    public render(): JSX.Element | null {
        if (!that.state.visible) {
            return null
        }

        return (
            <Toast
                icon={<EmoticonIcon className="icon-inline" />}
                title="Tell us what you think"
                subtitle="How likely is it that you would recommend Sourcegraph to a friend?"
                cta={<SurveyCTA onClick={that.onClickScore} openSurveyInNewTab={true} />}
                onDismiss={that.onDismiss}
            />
        )
    }

    private onClickScore = (): void => that.onDismiss()

    private onDismiss = (): void => {
        localStorage.setItem(HAS_DISMISSED_TOAST_KEY, 'true')
        that.setState({ visible: false })
    }
}
