import EmojiIcon from '@sourcegraph/icons/lib/Emoji'
import * as React from 'react'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../auth'
import { events } from '../tracking/events'
import { Toast } from './Toast'
import { daysActiveCount } from './util'

const HUBSPOT_SURVEY_URL = 'https://sourcegraph-2762526.hs-sites.com/user-survey'
const HAS_DISMISSED_TOAST_KEY = 'has-dismissed-survey-toast'

interface State {
    user: GQL.IUser | null
    visible: boolean
}

export class SurveyToast extends React.Component<{}, State> {
    private subscriptions = new Subscription()

    constructor() {
        super()
        this.state = {
            user: null,
            visible: localStorage.getItem(HAS_DISMISSED_TOAST_KEY) !== 'true' &&
                daysActiveCount === 3,
        }
        if (this.state.visible) {
            events.SurveyReminderViewed.log({ marketing: { sessionCount: daysActiveCount } })
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            currentUser.subscribe(
                user => this.setState({ user }),
            ),
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.visible) {
            return null
        }

        return (
            <Toast
                icon={<EmojiIcon className='icon-inline' />}
                title='Tell us what you think'
                subtitle='How likely is it that you would recommend Sourcegraph to a friend?'
                cta={
                    <div>{
                        Array(11).fill(1).map((_, i) =>
                            // tslint:disable-next-line:jsx-no-lambda
                            <button type='button' key={i} className='btn btn-primary toast__rating-btn' onClick={() => this.onClickSurvey(i)}>{i}</button>)
                    }</div>
                }
                onDismiss={this.onDismiss}
            />
        )
    }

    private onClickSurvey = (score: number): void => {
        events.SurveyReminderButtonClicked.log({ marketing: { nps_score: score } })
        const url = new URL(HUBSPOT_SURVEY_URL)
        url.searchParams.set('nps_score', score.toString())
        url.searchParams.set('user_is_authenticated', (this.state.user !== null).toString())
        if (this.state.user) {
            url.searchParams.set('email', this.state.user.email)
        }
        window.open(url.href)

        this.onDismiss()
    }

    private onDismiss = (): void => {
        localStorage.setItem(HAS_DISMISSED_TOAST_KEY, 'true')
        this.setState({ visible: false })
    }
}
