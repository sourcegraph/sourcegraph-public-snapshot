import { storiesOf } from '@storybook/react'
import React from 'react'
import { CampaignStatus } from './CampaignStatus'
import { BackgroundProcessState } from '../../../../../shared/src/graphql/schema'
import '../../../main.scss'
import { action } from '@storybook/addon-actions'
import { boolean } from '@storybook/addon-knobs'

const { add } = storiesOf('CampaignStatus', module).addDecorator(story => (
    <div className="theme-light container">{story()}</div>
))

add('Errored', () => (
    <CampaignStatus
        campaign={{
            viewerCanAdminister: boolean('Viewer can administer', true),
            status: {
                state: BackgroundProcessState.ERRORED,
                completedCount: 0,
                pendingCount: 2,
                errors: Array.from({ length: 3 }).map(
                    () =>
                        'creating commit from patch for repository "github.com/go-macaron/switcher": gitserver: pushing ref: exit status 128\n' +
                        '```\n' +
                        '$ git push --force https://@github.com/go-macaron/switcher b1275f86053a021c630b354b414e522ce73244a1:refs/heads/campaign/migrate-from-travis-to-github-actions\n' +
                        'remote: Permission to go-macaron/switcher.git denied to foobar.\n' +
                        "fatal: unable to access 'https://github.com/go-macaron/switcher': The requested URL returned error: 403\n" +
                        '```'
                ),
            },
            changesets: {
                totalCount: 0,
            },
            publishedAt: boolean('Is draft', false) ? null : new Date().toISOString(),
            closedAt: null,
        }}
        onPublish={action('Publish')}
        onRetry={action('Retry')}
    />
))
