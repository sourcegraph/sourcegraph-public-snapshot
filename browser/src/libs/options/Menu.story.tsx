import * as React from 'react'

import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'

import '../../app.scss'

import { OptionsMenu } from './Menu'

// tslint:disable: jsx-no-lambda

storiesOf('Options - OptionsMenu', module)
    .add('Default', () => (
        <div style={{ maxWidth: 400, marginLeft: 20, marginTop: 20, boxShadow: '0 0 12px 0 rgba(0, 0, 0, 0.15)' }}>
            <OptionsMenu
                version={'0.0.0'}
                status={'connected'}
                sourcegraphURL={'https://sourcegraph.com'}
                onURLChange={action('Sourcegraph URL changed')}
                onURLSubmit={action('New Sourcegraph URL submitted')}
                onSettingsClick={action('Settings clicked')}
                featureFlags={[]}
                isSettingsOpen={false}
                toggleFeatureFlag={action('Feature flag toggled')}
                requestPermissions={() => void 0}
                urlHasPermissions={true}
            />
        </div>
    ))
    .add('Settings open', () => (
        <div style={{ maxWidth: 400, marginLeft: 20, marginTop: 20, boxShadow: '0 0 12px 0 rgba(0, 0, 0, 0.15)' }}>
            <OptionsMenu
                version={'0.0.0'}
                status={'connected'}
                sourcegraphURL={'https://sourcegraph.com'}
                onURLChange={action('Sourcegraph URL changed')}
                onURLSubmit={action('New Sourcegraph URL submitted')}
                onSettingsClick={action('Settings clicked')}
                isSettingsOpen={true}
                toggleFeatureFlag={action('Feature flag toggled')}
                requestPermissions={() => void 0}
                urlHasPermissions={true}
            />
        </div>
    ))
