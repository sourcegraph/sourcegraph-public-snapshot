import * as React from 'react'
import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import { OptionsMenu } from './OptionsMenu'
import optionsStyles from '../../options.scss'

storiesOf('browser/Options/OptionsMenu', module)
    .addDecorator(story => (
        <>
            <style>{optionsStyles}</style>
            <div>{story()}</div>
        </>
    ))
    .add('Default', () => (
        <OptionsMenu
            version="0.0.0"
            status="connected"
            sourcegraphURL="https://sourcegraph.com"
            onURLChange={action('Sourcegraph URL changed')}
            onURLSubmit={action('New Sourcegraph URL submitted')}
            onSettingsClick={action('Settings clicked')}
            onToggleActivationClick={action('Toggle activation clicked')}
            featureFlags={[]}
            isSettingsOpen={false}
            isActivated={true}
            toggleFeatureFlag={action('Feature flag toggled')}
            requestPermissions={() => undefined}
            urlHasPermissions={true}
        />
    ))
    .add('Settings open', () => (
        <OptionsMenu
            version="0.0.0"
            status="connected"
            sourcegraphURL="https://sourcegraph.com"
            onURLChange={action('Sourcegraph URL changed')}
            onURLSubmit={action('New Sourcegraph URL submitted')}
            onSettingsClick={action('Settings clicked')}
            onToggleActivationClick={action('Toggle activation clicked')}
            isSettingsOpen={true}
            isActivated={true}
            toggleFeatureFlag={action('Feature flag toggled')}
            featureFlags={[
                { key: 'Test setting 1', value: true },
                { key: 'Test setting 2', value: false },
            ]}
            requestPermissions={() => undefined}
            urlHasPermissions={true}
        />
    ))
