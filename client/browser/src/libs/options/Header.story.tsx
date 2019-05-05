import * as React from 'react'

import { storiesOf } from '@storybook/react'

import '../../app.scss'

import { action } from '@storybook/addon-actions'
import { OptionsHeader } from './Header'

storiesOf('Options - OptionsHeader', module).add('Default', () => (
    <div style={{ maxWidth: 400 }}>
        <OptionsHeader version={'0.0.0'} onSettingsClick={action('Settings clicked')} />
    </div>
))
