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
            onClickExpandOptionsMenu={action('Expand clicked')}
            onToggleActivationClick={action('Toggle activation clicked')}
            optionFlags={[]}
            isOptionsMenuExpanded={false}
            isActivated={true}
            onChangeOptionFlag={action('Feature flag changed')}
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
            isOptionsMenuExpanded={true}
            isActivated={true}
            onChangeOptionFlag={action('Feature flag changed')}
            optionFlags={[
                { label: 'Test setting 1', key: 'testSetting1', value: true },
                { label: 'Test setting 2', key: 'testSetting2', value: false },
            ]}
            requestPermissions={() => undefined}
            urlHasPermissions={true}
        />
    ))
