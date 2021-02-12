import React from 'react'
import { TabsWithLocalStorageViewStatePersistence } from './Tabs'
import { mount } from 'enzyme'

describe('Tab', () => {
    test('Tabs with local storage persistenc', () => {
        expect(
            mount(
                <TabsWithLocalStorageViewStatePersistence
                    tabs={[{ id: 'files', label: 'Files' }]}
                    storageKey="repo-revision-sidebar-last-tab"
                    tabBarEndFragment={
                        <>
                            <div>fragment</div>
                        </>
                    }
                    id="tab"
                    className="tab-bar"
                    tabClassName="tab-bar__tab--h5like"
                    onSelectTab={() => null}
                >
                    <div>Panel 1</div>
                    <div>Panel 2</div>
                </TabsWithLocalStorageViewStatePersistence>
            )
        ).toMatchSnapshot()
    })
})
