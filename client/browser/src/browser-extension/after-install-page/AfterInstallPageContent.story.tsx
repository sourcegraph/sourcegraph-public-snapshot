import { storiesOf } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'

import brandedStyles from '../../branded.scss'

import { AfterInstallPageContent } from './AfterInstallPageContent'

storiesOf('browser/AfterInstallPage', module).add('Default', () => (
    <BrandedStory styles={brandedStyles}>{AfterInstallPageContent}</BrandedStory>
))
