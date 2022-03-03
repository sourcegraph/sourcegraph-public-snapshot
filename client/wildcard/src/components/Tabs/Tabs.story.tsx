import { Meta, Story } from '@storybook/react'
import React from 'react'

import brandedStyles from '@sourcegraph/branded/src/global-styles/index.scss'
import { usePrependStyles } from '@sourcegraph/storybook/src/hooks/usePrependStyles'

import { Tabs, Tab, TabList, TabPanel, TabPanels, TabsProps } from '.'

export const TabsStory: Story<TabsProps & { actions: boolean }> = args => {
    usePrependStyles('branded-story-styles', brandedStyles)

    const { actions, lazy, behavior, size, ...props } = args

    return (
        <Tabs lazy={lazy} behavior={behavior} size={size} {...props}>
            <TabList actions={actions ? <div>custom component rendered</div> : null}>
                <Tab>Tab 1</Tab>
                <Tab>Tab 2</Tab>
            </TabList>
            <TabPanels>
                <TabPanel>Panel 1</TabPanel>
                <TabPanel>Panel 2</TabPanel>
            </TabPanels>
        </Tabs>
    )
}

TabsStory.storyName = 'Tabs component'

const config: Meta = {
    title: 'wildcard/Tabs',
    component: Tabs,
    parameters: {
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
        design: [
            {
                type: 'figma',
                name: 'Figma Light',
                url:
                    'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=954%3A5153',
            },
            {
                type: 'figma',
                name: 'Figma Dark',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=954%3A6125',
            },
        ],
    },
    argTypes: {
        size: {
            options: ['small', 'medium', 'large'],
            control: { type: 'radio' },
        },
        lazy: {
            options: [true, false],
            control: { type: 'radio' },
        },
        behavior: {
            options: ['memoize', 'forceRender'],
            control: { type: 'radio' },
        },
        actions: {
            options: [true, false],
            control: { type: 'radio' },
        },
    },
}

export default config
