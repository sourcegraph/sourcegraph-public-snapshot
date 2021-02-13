import React from 'react'
import { Tabs, TabList, Tab, TabPanels, TabPanel } from '@reach/tabs'

export const Tabs1: React.FunctionComponent<{}> = () => (
    <Tabs>
        <TabList>
            <Tab>One</Tab>
            <Tab>Two</Tab>
            <Tab>Three</Tab>
        </TabList>
        <TabPanels>
            <TabPanel>
                <p>one!</p>
            </TabPanel>
            <TabPanel>
                <p>two!</p>
            </TabPanel>
            <TabPanel>
                <p>three!</p>
            </TabPanel>
        </TabPanels>
    </Tabs>
)
