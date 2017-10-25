import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'
import reactive from 'rx-component'
import 'rxjs/add/observable/merge'
import 'rxjs/add/observable/of'
import 'rxjs/add/operator/concat'
import 'rxjs/add/operator/concat'
import 'rxjs/add/operator/delay'
import 'rxjs/add/operator/do'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/scan'
import 'rxjs/add/operator/startWith'
import 'rxjs/add/operator/withLatestFrom'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { updateOrg } from '../backend'

export interface Props {
    org: GQL.IOrg
}

interface State {
    orgID: number
    slackWebhookURL: string
    displayName: string
    loading: boolean
    updated: boolean
}

type Update = (state: State) => State

export const OrgSettingsForm = reactive<Props>(props => {

    const submits = new Subject<React.FormEvent<HTMLFormElement>>()
    const nextSubmit = (e: React.FormEvent<HTMLFormElement>) => submits.next(e)

    const displayNameChanges = new Subject<string>()
    const nextDisplayNameChange = (event: React.ChangeEvent<HTMLInputElement>) => displayNameChanges.next(event.currentTarget.value)

    const slackWebhookURLChanges = new Subject<string>()
    const nextSlackWebhookURLChange = (event: React.ChangeEvent<HTMLInputElement>) => slackWebhookURLChanges.next(event.currentTarget.value)

    const orgID = props.map(({ org }) => org.id)
    const displayNames = Observable.merge(
        props.map(props => props.org.displayName || ''),
        displayNameChanges
    )
    const slackWebhookURLs = Observable.merge(
        props.map(props => props.org.slackWebhookURL || ''),
        slackWebhookURLChanges
    )

    return Observable.merge<Update>(
        orgID
            .map(orgID => (state: State): State => ({ ...state, orgID })),

        displayNames
            .map(displayName => (state: State): State => ({ ...state, displayName })),

        slackWebhookURLs
            .map(slackWebhookURL => (state: State): State => ({ ...state, slackWebhookURL })),

        submits
            .do(e => e.preventDefault())
            .withLatestFrom(orgID, displayNames, slackWebhookURLs)
            .mergeMap(([, orgID, displayName, slackWebhookURL]) =>
                updateOrg(orgID, displayName, slackWebhookURL)
                    .mergeMap(() =>
                        // Reset email, reenable submit button, flash "updated" text
                        Observable.of((state: State): State => ({ ...state, loading: false, updated: true }))
                            // Hide "updated" text again after 1s
                            .concat(Observable.of<Update>(state => ({ ...state, updated: false })).delay(1000))
                    )
                    // Disable button while loading
                    .startWith<Update>((state: State): State => ({ ...state, loading: true }))
            )
    )
        .scan<Update, State>((state: State, update: Update) => update(state), { updated: false, loading: false, slackWebhookURL: '', displayName: '' } as State)
        .map(({ loading, slackWebhookURL, displayName, updated }) => (
            <form className='org-settings-form' onSubmit={nextSubmit}>
                <h3>Organization settings</h3>
                <div className='form-group'>
                    <label>Display name</label>
                    <input
                        type='text'
                        className='ui-text-box org-settings-form__display-name'
                        placeholder='Organization name'
                        onChange={nextDisplayNameChange}
                        value={displayName}
                        spellCheck={false}
                        size={60}
                    />
                </div>
                <div className='form-group'>
                    <label>Slack webhook URL</label>
                    <input
                        type='url'
                        className='ui-text-box org-settings-form__slack-webhook-url'
                        placeholder=''
                        onChange={nextSlackWebhookURLChange}
                        value={slackWebhookURL}
                        spellCheck={false}
                        size={60}
                    />
                    <small className='form-text'>
                        Integrate Sourcegraph's code comments and org updates with your team's Slack channel!
                Visit &lt;your-workspace-url&gt;.slack.com/apps/manage/custom-integrations > Incoming Webhooks > Add Configuration.
                </small>
                </div>
                <button type='submit' disabled={loading} className='btn btn-primary org-settings-form__submit-button'>Update</button>
                {loading && <LoaderIcon className='icon-inline' />}
                <div className={'org-settings-form__updated-text' + (updated ? ' org-settings-form__updated-text--visible' : '')}><small>Updated!</small></div>
            </form>
        ))
})
