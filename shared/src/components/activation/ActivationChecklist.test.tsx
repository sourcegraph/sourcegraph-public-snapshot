import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { ActivationChecklist } from './ActivationChecklist'

describe('ActivationChecklist', () => {
    test('render loading', () => {
        const component = renderer.create(
            <ActivationChecklist steps={[]} history={H.createMemoryHistory({ keyLength: 0 })} />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })
    test('render 0/1 complete', () => {
        {
            const component = renderer.create(
                <ActivationChecklist
                    steps={[
                        {
                            id: 'item1',
                            title: 'title1',
                            detail: 'detail1',
                            action: () => undefined,
                        },
                    ]}
                    completed={{}}
                    history={H.createMemoryHistory({ keyLength: 0 })}
                />
            )
            expect(component.toJSON()).toMatchSnapshot()
        }
        {
            const component = renderer.create(
                <ActivationChecklist
                    steps={[
                        {
                            id: 'item1',
                            title: 'title1',
                            detail: 'detail1',
                            action: () => undefined,
                        },
                    ]}
                    completed={{ anotherItem: true }}
                    history={H.createMemoryHistory({ keyLength: 0 })}
                />
            )
            expect(component.toJSON()).toMatchSnapshot()
        }
    })
    test('render 1/1 complete', () => {
        const component = renderer.create(
            <ActivationChecklist
                steps={[
                    {
                        id: 'item1',
                        title: 'title1',
                        detail: 'detail1',
                        action: () => undefined,
                    },
                ]}
                completed={{ item1: true }}
                history={H.createMemoryHistory({ keyLength: 0 })}
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })
})
