import type { Meta, StoryFn } from '@storybook/react'

import { H1, H2 } from '..'
import { BrandedStory } from '../../stories/BrandedStory'

import { Tabs, Tab, TabList, TabPanel, TabPanels, type TabsProps } from '.'

export const TabsStory: StoryFn<TabsProps & { actions: boolean }> = args => (
    <>
        <H1>Tabs</H1>
        <Container title="Standard">
            <TabsVariant {...args} />
        </Container>
        <Container width={300} title="Limited width">
            <TabsVariant {...args} />
        </Container>
        <Container width={300} title="Scrolled tab list">
            <TabsVariant {...args} longTabList="scroll" />
        </Container>
    </>
)

TabsStory.storyName = 'Tabs component'

const config: Meta = {
    title: 'wildcard/Tabs',
    component: Tabs,
    decorators: [story => <BrandedStory>{() => story()}</BrandedStory>],
    parameters: {
        design: [
            {
                type: 'figma',
                name: 'Figma Light',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=954%3A5153',
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

const TabsVariant: StoryFn<TabsProps & { actions: boolean }> = args => {
    const { actions, lazy, behavior, size, ...props } = args
    return (
        <Tabs lazy={lazy} behavior={behavior} size={size} {...props}>
            <TabList actions={actions ? <div>custom component rendered</div> : null}>
                <Tab>Tab 1</Tab>
                <Tab>Tab 2</Tab>
                <Tab>Third tab</Tab>
                <Tab>Fourth tab</Tab>
                <Tab>Fifth tab</Tab>
                <Tab>Sixth tab</Tab>
            </TabList>
            <TabPanels>
                <TabPanel>Panel 1</TabPanel>
                <TabPanel>Panel 2</TabPanel>
                <TabPanel>Panel 3</TabPanel>
                <TabPanel>Panel 4</TabPanel>
                <TabPanel>Panel 5</TabPanel>
                <TabPanel>Panel 6</TabPanel>
            </TabPanels>
        </Tabs>
    )
}

interface ContainerProps {
    title: string
    width?: number
}

const Container: React.FunctionComponent<React.PropsWithChildren<ContainerProps>> = ({ title, width, children }) => (
    <>
        <H2 style={{ margin: '30px 0 10px 0' }}>{title}</H2>
        <div style={{ width: width ? `${width}px` : undefined }}>{children}</div>
    </>
)

export default config
