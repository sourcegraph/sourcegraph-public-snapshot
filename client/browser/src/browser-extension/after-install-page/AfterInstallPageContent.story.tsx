import React from 'react'
import { storiesOf } from '@storybook/react'
import { AfterInstallPageContent } from './AfterInstallPageContent'
import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import brandedStyles from '../../branded.scss'

storiesOf('browser/AfterInstallPage', module).add('Default', () => (
    <BrandedStory styles={brandedStyles}>{AfterInstallPageContent}</BrandedStory>
))
