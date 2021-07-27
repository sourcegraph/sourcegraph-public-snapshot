import { Meta, Story } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'

import { Container } from '..'

import { Tabs, Tab, TabList, TabPanel, TabPanels, TabsProps } from '.'

export const TabsStory: Story<TabsProps & { actions: boolean }> = args => {
    const { actions, ...props } = args

    return (
        <BrandedStory>
            {() => (
                <Container>
                    <Tabs {...props}>
                        <TabList actions={actions ? <div>custom component rendered</div> : null}>
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
}

TabsStory.storyName = 'Tabs component'

const config: Meta = {
    title: 'wildcard/Tabs',
    component: Tabs,
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

// eslint-disable-next-line import/no-default-export
export default config
