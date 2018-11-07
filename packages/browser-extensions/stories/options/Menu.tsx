import * as React from 'react'

import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'

import '../global.scss'

import { OptionsMenu } from '../../src/libs/options/Menu'

storiesOf('Options - OptionsMenu', module)
    .add('Default', () => (
        <div style={{ maxWidth: 400, marginLeft: 20, marginTop: 20, boxShadow: '0 0 12px 0 rgba(0, 0, 0, 0.15)' }}>
            <OptionsMenu
                version={'0.0.0'}
                onSettingsClick={action('Settings click')}
                isSettingsOpen={false}
                status={'connected'}
                sourcegraphURL={'https://sourcegraph.com'}
                onURLChange={action('Sourcegraph URL changed')}
                onURLSubmit={action('New Sourcegraph URL submitted')}
                settings={{ hello: 'World', nested: { object: true } }}
                onSettingsChange={action('Settings Changed')}
                onSettingsSave={action('Settings Saved')}
                settingsHaveChanged={true}
            />
        </div>
    ))
    .add('Settings open', () => (
        <div style={{ maxWidth: 400, marginLeft: 20, marginTop: 20, boxShadow: '0 0 12px 0 rgba(0, 0, 0, 0.15)' }}>
            <OptionsMenu
                version={'0.0.0'}
                onSettingsClick={action('Settings click')}
                isSettingsOpen={true}
                status={'connected'}
                sourcegraphURL={'https://sourcegraph.com'}
                onURLChange={action('Sourcegraph URL changed')}
                onURLSubmit={action('New Sourcegraph URL submitted')}
                settings={{ hello: 'World', nested: { object: true } }}
                onSettingsChange={action('Settings Changed')}
                onSettingsSave={action('Settings Saved')}
                settingsHaveChanged={true}
            />
        </div>
    ))
