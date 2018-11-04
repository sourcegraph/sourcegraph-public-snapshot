import * as React from 'react'

import { storiesOf } from '@storybook/react'

import '../global.scss'

import { ConfigWarning } from '../../src/shared/components/options/ConfigWarning'

storiesOf('ConnectionCard', module).add('Default', () => <ConfigWarning />)
