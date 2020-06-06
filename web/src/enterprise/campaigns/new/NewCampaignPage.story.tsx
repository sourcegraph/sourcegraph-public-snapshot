import { storiesOf } from '@storybook/react'
import React from 'react'
import webStyles from '../../../SourcegraphWebApp.scss'
import * as H from 'history'
import { NewCampaignPage } from './NewCampaignPage'
import { throwError, timer } from 'rxjs'
import { mergeMapTo } from 'rxjs/operators'

const history = H.createMemoryHistory()

const { add } = storiesOf('web/campaigns/NewCampaignPage', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="theme-light container mt-3">{story()}</div>
    </>
))

add('Empty', () => (
    <NewCampaignPage
        authenticatedUser={{ id: 'u', username: 'alice', avatarURL: null }}
        history={history}
        _createCampaign={() =>
            timer(1000)
                .pipe(mergeMapTo(throwError(new Error('x'))))
                .toPromise()
        }
    />
))
