import { Meta, Story } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Container } from '..'

import { Tabs, Tab, TabList, TabPanel, TabPanels } from '.'

export const TabsStory: Story = () => (
    <BrandedStory styles={webStyles}>
        {() => (
            <Container>
                <Tabs lazy={true} behavior="memoize" size="large">
                    <TabList actions={<div>custom component rendered</div>}>
                        <Tab>Tab 1</Tab>
                        <Tab>Tab 2</Tab>
                    </TabList>
                    <TabPanels>
                        <TabPanel>Panel 1</TabPanel>
                        <TabPanel>Panel 2</TabPanel>
                    </TabPanels>
                </Tabs>
            </Container>
        )}
    </BrandedStory>
)

TabsStory.storyName = 'Tabs component'

// eslint-disable-next-line import/no-default-export
export default {
    title: 'wildcard/Tabs',
    component: TabsStory,
    args: {
        size: ['small', 'medium', 'large'],
    },
} as Meta
