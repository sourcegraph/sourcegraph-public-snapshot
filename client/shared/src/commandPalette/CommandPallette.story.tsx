import { Meta, Story } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { CommandPalette, CommandPaletteProps } from './CommandPalette'

const config: Meta = {
    title: 'shared/CommandPallette',

    parameters: {
        component: CommandPalette,
    },
}

export default config

export const Default: Story<CommandPaletteProps> = () => (
    <BrandedStory styles={webStyles}>{() => <CommandPalette />}</BrandedStory>
)
