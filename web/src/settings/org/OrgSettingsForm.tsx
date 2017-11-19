import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'
import reactive from 'rx-component'
import { merge } from 'rxjs/observable/merge'
import { of } from 'rxjs/observable/of'
import { concat } from 'rxjs/operators/concat'
import { delay } from 'rxjs/operators/delay'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { scan } from 'rxjs/operators/scan'
import { startWith } from 'rxjs/operators/startWith'
import { tap } from 'rxjs/operators/tap'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { Subject } from 'rxjs/Subject'
import { updateOrg } from '../backend'

export interface Props {
    org: GQL.IOrg
}

interface State {
    orgID: string
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
    const nextDisplayNameChange = (event: React.ChangeEvent<HTMLInputElement>) =>
        displayNameChanges.next(event.currentTarget.value)

    const slackWebhookURLChanges = new Subject<string>()
    const nextSlackWebhookURLChange = (event: React.ChangeEvent<HTMLInputElement>) =>
        slackWebhookURLChanges.next(event.currentTarget.value)

    const orgID = props.pipe(map(({ org }) => org.id))
    const displayNames = merge(props.pipe(map(props => props.org.displayName || '')), displayNameChanges)
    const slackWebhookURLs = merge(props.pipe(map(props => props.org.slackWebhookURL || '')), slackWebhookURLChanges)

    return merge<Update>(
        orgID.pipe(map(orgID => (state: State): State => ({ ...state, orgID }))),

        displayNames.pipe(map(displayName => (state: State): State => ({ ...state, displayName }))),

        slackWebhookURLs.pipe(map(slackWebhookURL => (state: State): State => ({ ...state, slackWebhookURL }))),

        submits.pipe(
            tap(e => e.preventDefault()),
            withLatestFrom(orgID, displayNames, slackWebhookURLs),
            mergeMap(([, orgID, displayName, slackWebhookURL]) =>
                updateOrg(orgID, displayName, slackWebhookURL).pipe(
                    mergeMap(() =>
                        // Reset email, reenable submit button, flash "updated" text
                        of((state: State): State => ({ ...state, loading: false, updated: true }))
                            // Hide "updated" text again after 1s
                            .pipe(concat(of<Update>(state => ({ ...state, updated: false })).pipe(delay(1000))))
                    ),
                    // Disable button while loading
                    startWith<Update>((state: State): State => ({ ...state, loading: true }))
                )
            )
        )
    ).pipe(
        scan<Update, State>((state: State, update: Update) => update(state), {
            updated: false,
            loading: false,
            slackWebhookURL: '',
            displayName: '',
        } as State),
        map(({ loading, slackWebhookURL, displayName, updated }) => (
            <form className="org-settings-form" onSubmit={nextSubmit}>
                <h3>Organization settings</h3>
                <div className="form-group">
                    <label>Display name</label>
                    <input
                        type="text"
                        className="ui-text-box org-settings-form__display-name"
                        placeholder="Organization name"
                        onChange={nextDisplayNameChange}
                        value={displayName}
                        spellCheck={false}
                        size={60}
                    />
                </div>
                <div className="form-group">
                    <label>Slack webhook URL</label>
                    <input
                        type="url"
                        className="ui-text-box org-settings-form__slack-webhook-url"
                        placeholder=""
                        onChange={nextSlackWebhookURLChange}
                        value={slackWebhookURL}
                        spellCheck={false}
                        size={60}
                    />
                    <small className="form-text">
                        Integrate Sourcegraph's code comments and org updates with your team's Slack channel! Visit
                        &lt;your-workspace-url&gt;.slack.com/apps/manage/custom-integrations > Incoming Webhooks > Add
                        Configuration.
                    </small>
                </div>
                <button type="submit" disabled={loading} className="btn btn-primary org-settings-form__submit-button">
                    Update
                </button>
                {loading && <LoaderIcon className="icon-inline" />}
                <div
                    className={
                        'org-settings-form__updated-text' + (updated ? ' org-settings-form__updated-text--visible' : '')
                    }
                >
                    <small>Updated!</small>
                </div>
            </form>
        ))
    )
})
