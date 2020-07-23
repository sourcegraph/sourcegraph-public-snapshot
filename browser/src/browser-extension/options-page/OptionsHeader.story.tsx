import * as React from 'react'
import { storiesOf } from '@storybook/react'
import { action } from '@storybook/addon-actions'
import { OptionsHeader } from './OptionsHeader'
import optionsStyles from '../../options.scss'

storiesOf('browser/Options/OptionsHeader', module)
    .addDecorator(story => (
        <>
            <style>{optionsStyles}</style>
            <div>{story()}</div>
        </>
    ))
    .add('Default', () => (
        <div style={{ maxWidth: 400 }}>
            <OptionsHeader
                version="0.0.0"
                isActivated={true}
                onSettingsClick={action('Settings clicked')}
                onToggleActivationClick={action('Toggle activation clicked')}
            />
        </div>
    ))
