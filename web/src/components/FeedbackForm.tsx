import CloseIcon from '@sourcegraph/icons/lib/Close'
import EmoticonIcon from '@sourcegraph/icons/lib/Emoticon'
import EmoticonSadIcon from '@sourcegraph/icons/lib/EmoticonSad'
import TwitterIcon from '@sourcegraph/icons/lib/Twitter'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { fromEvent, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'
import * as GQL from '../backend/graphqlschema'
import { eventLogger } from '../tracking/eventLogger'
import { pluralize } from '../util/strings'
import { Form } from './Form'

interface Props {
    user: GQL.IUser | null
    onDismiss: () => void
}

type Experience = 'good' | 'bad'
interface State {
    experience?: Experience
    description: string
}

const DESCRIPTION_LOCAL_STORAGE_KEY = 'FeedbackForm-description'
const EXPERIENCE_LOCAL_STORAGE_KEY = 'FeedbackForm-experience'
const ISSUES_URL = 'https://github.com/sourcegraph/issues'
const TWITTER_URL = 'https://twitter.com/intent/tweet?'
const TWEET_HASHTAG = ' #UseTheSource'
const TWEET_MENTION = ' via @srcgraph'

export class FeedbackForm extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = {
            description: localStorage.getItem(DESCRIPTION_LOCAL_STORAGE_KEY) || '',
            experience: (localStorage.getItem(EXPERIENCE_LOCAL_STORAGE_KEY) as Experience | null) || undefined,
        }
    }

    public componentDidMount(): void {
        // Hide feedback form if escape key is pressed and text field isn't focused.
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(filter(event => event.key === 'Escape'))
                .subscribe(() => this.props.onDismiss())
        )
    }

    public componentDidUpdate(): void {
        localStorage.setItem(DESCRIPTION_LOCAL_STORAGE_KEY, this.state.description + '')
        localStorage.setItem(EXPERIENCE_LOCAL_STORAGE_KEY, this.state.experience + '')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element {
        return (
            <Form className="feedback-form card" onSubmit={this.handleSubmit}>
                <div className="card-body">
                    <button type="reset" className="btn btn-icon feedback-form__close" onClick={this.props.onDismiss}>
                        <CloseIcon />
                    </button>
                    <h2>Tweet us your feedback</h2>
                    <div className="form-group">
                        <label>How was your experience?</label>
                        <div>
                            <button
                                type="button"
                                className={
                                    'btn btn-icon feedback-form__emoticon' +
                                    (this.state.experience === 'good' ? ' feedback-form__emoticon--happy' : '')
                                }
                                onClick={this.saveGoodExperience}
                            >
                                <EmoticonIcon />
                            </button>
                            <button
                                type="button"
                                className={
                                    'btn btn-icon feedback-form__emoticon' +
                                    (this.state.experience === 'bad' ? ' feedback-form__emoticon--sad' : '')
                                }
                                onClick={this.saveBadExperience}
                            >
                                <EmoticonSadIcon />
                            </button>
                        </div>
                    </div>
                    <div className="form-group">
                        <label>Tell us why:</label>{' '}
                        <small className="text-muted">
                            {this.calculateMaxCharacters() - this.state.description.length}{' '}
                            {pluralize('character', this.calculateMaxCharacters() - this.state.description.length)} left
                        </small>
                        <textarea
                            name="description"
                            id="description"
                            className="form-control"
                            onChange={this.handleDescriptionChange}
                            value={this.state.description}
                            required={true}
                            maxLength={this.calculateMaxCharacters()}
                            autoFocus={true}
                        />
                    </div>
                    <div className="form-group">
                        <button type="submit" className="btn btn-primary">
                            <TwitterIcon className="icon icon-inline" /> Tweet us
                        </button>{' '}
                        <button type="reset" className="btn btn-secondary" onClick={this.props.onDismiss}>
                            Cancel
                        </button>
                    </div>
                    <div>
                        Or{' '}
                        <Link to={ISSUES_URL} onClick={this.reportIssue} target="_blank">
                            report an issue
                        </Link>.
                    </div>
                </div>
            </Form>
        )
    }

    /**
     * Tells if the query is unsupported for sending notifications.
     */

    private reportIssue = () => {
        eventLogger.log('ReportIssueButtonClicked')
    }
    private handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault()

        const url = new URL(TWITTER_URL)

        let experienceEmoji = ''

        if (this.state.experience === 'good') {
            experienceEmoji = ' 😄'
        }

        if (this.state.experience === 'bad') {
            experienceEmoji = ' 😞'
        }

        url.searchParams.set(
            'text',
            this.state.description +
                experienceEmoji +
                (this.state.experience === 'good' ? TWEET_HASHTAG : '') +
                TWEET_MENTION
        )

        window.open(url.href)

        eventLogger.log('FeedbackSubmitted', {
            feedback: {
                experience: this.state.experience ? this.state.experience : undefined,
            },
        })

        localStorage.removeItem(DESCRIPTION_LOCAL_STORAGE_KEY)
        localStorage.removeItem(EXPERIENCE_LOCAL_STORAGE_KEY)
        this.props.onDismiss()
    }
    /**
     * Calculates max characters for the description field
     */
    private calculateMaxCharacters(): number {
        let maxCharacters = 280 - TWEET_MENTION.length

        if (this.state.experience === 'good') {
            maxCharacters -= (' 😄' + TWEET_HASHTAG).length
        } else if (this.state.experience === 'bad') {
            maxCharacters -= ' 😞'.length
        }
        return maxCharacters
    }

    private saveGoodExperience = (): void => {
        this.setState({ experience: 'good' })
        eventLogger.log('FeedbackGoodExperienceClicked')
    }

    private saveBadExperience = (): void => {
        this.setState({ experience: 'bad' })
        eventLogger.log('FeedbackBadExperienceClicked')
    }
    /**
     * Handles description change by updating the component's state
     */
    private handleDescriptionChange = (event: React.FocusEvent<HTMLTextAreaElement>): void => {
        this.setState({ description: event.currentTarget.value })
    }
}
