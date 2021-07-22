import { render, RenderResult, cleanup, fireEvent } from '@testing-library/react'
import React from 'react'

import { Tab, TabList, TabPanel, TabPanels, Tabs, TabsProps } from './Tabs'

const TabsTest = (props: TabsProps) => <Tabs {...props} />

const TabsChildren = () => (
    <>
        <TabList>
            <Tab>Tab 1</Tab>
            <Tab>Tab 2</Tab>
        </TabList>
        <TabPanels>
            <TabPanel forceRender={true}>Panel 1</TabPanel>
            <TabPanel forceRender={true}>Panel 2</TabPanel>
        </TabPanels>
    </>
)

const TabsNoForceRender = () => (
    <>
        <TabList>
            <Tab>Tab 1</Tab>
            <Tab>Tab 2</Tab>
        </TabList>
        <TabPanels>
            <TabPanel forceRender={false}>Panel 1</TabPanel>
            <TabPanel forceRender={false}>Panel 2</TabPanel>
        </TabPanels>
    </>
)

describe('Tabs', () => {
    let queries: RenderResult

    const renderWithProps = (props: TabsProps): RenderResult => render(<TabsTest {...props} />)

    afterEach(cleanup)

    describe('Invalid configuration', () => {
        it('will error when no children are added', () => {
            expect(() => {
                renderWithProps({ index: 1, children: undefined })
            }).toThrowErrorMatchingSnapshot()
        })
    })

    describe('Tabs with forceRender=true', () => {
        beforeEach(() => {
            queries = renderWithProps({ children: <TabsChildren /> })
        })
        it('will render tabs children correctly', () => {
            expect(queries.getByTestId('wildcard-tabs')).toBeInTheDocument()
            expect(queries.getByTestId('wildcard-tab-list')).toBeInTheDocument()
            expect(queries.getByTestId('wildcard-tab-panels')).toBeInTheDocument()
        })
        it('will render the right amount of <Tab/> components', () => {
            const tabGroup = queries.getAllByTestId('wildcard-tab')
            tabGroup.forEach(tab => {
                expect(tab).toBeInTheDocument()
            })
            expect(queries.getAllByTestId('wildcard-tab')).toHaveLength(2)
        })
        it('will render the right amount of <TabPanel/> components', () => {
            const tabPanelGroup = queries.getAllByTestId('wildcard-tab')
            tabPanelGroup.forEach(tabPanelGroup => {
                expect(tabPanelGroup).toBeInTheDocument()
            })
            expect(queries.getAllByTestId('wildcard-tab-panel')).toHaveLength(2)
        })

        it('will render <TabPanel/> children each time associated <Tab>  is clicked', () => {
            fireEvent.click(queries.getAllByTestId('wildcard-tab')[0])
            expect(queries.getByText('Panel 1')).toBeInTheDocument()
            expect(queries.queryByText('Panel 2')).toBeNull()
            fireEvent.click(queries.getAllByTestId('wildcard-tab')[1])
            expect(queries.getByText('Panel 2')).toBeInTheDocument()
            expect(queries.queryByText('Panel 1')).toBeNull()
        })
    })

    describe('Tabs with forceRender=false', () => {
        beforeEach(() => {
            queries = renderWithProps({ children: <TabsNoForceRender /> })
        })

        it('will render <TabPanel/> children each time associated <Tab>  is clicked', () => {
            fireEvent.click(queries.getAllByTestId('wildcard-tab')[0])
            expect(queries.getByText('Panel 1')).toBeInTheDocument()
            expect(queries.queryByText('Panel 2')).toBeInTheDocument()
            fireEvent.click(queries.getAllByTestId('wildcard-tab')[1])
            expect(queries.getByText('Panel 2')).toBeInTheDocument()
            expect(queries.queryByText('Panel 1')).toBeInTheDocument()
        })
    })
})
