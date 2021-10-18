import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'

import brandedStyles from '../../branded.scss'

import { AfterInstallPageContent } from './AfterInstallPageContent'

export default {
    title: 'browser/AfterInstallPage',
}

export const Default = () => <BrandedStory styles={brandedStyles}>{AfterInstallPageContent}</BrandedStory>
