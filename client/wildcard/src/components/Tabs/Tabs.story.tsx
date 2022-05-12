import { Meta, Story } from '@storybook/react'

import brandedStyles from '@sourcegraph/branded/src/global-styles/index.scss'
import { usePrependStyles } from '@sourcegraph/storybook'

import { Typography } from '..'

import { Tabs, Tab, TabList, TabPanel, TabPanels, TabsProps } from '.'

export const TabsStory: Story<TabsProps & { actions: boolean }> = args => {
    usePrependStyles('branded-story-styles', brandedStyles)

    return (
        <>
            <Typography.H1>Tabs</Typography.H1>
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

const TabsVariant: Story<TabsProps & { actions: boolean }> = args => {
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

const Container: React.FunctionComponent<{ title: string; width?: number }> = ({ title, width, children }) => (
    <>
        <Typography.H2 style={{ margin: '30px 0 10px 0' }}>{title}</Typography.H2>
        <div style={{ width: width ? `${width}px` : undefined }}>{children}</div>
    </>
)

export default config
