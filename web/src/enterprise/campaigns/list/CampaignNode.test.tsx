import React from 'react'
import { CampaignNode } from './CampaignNode'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { parseISO } from 'date-fns'
import { createMemoryHistory } from 'history'
import { mount } from 'enzyme'

jest.mock('../icons', () => ({ CampaignsIcon: 'CampaignsIcon' }))

describe('CampaignNode', () => {
    const node = {
        __typename: 'Campaign',
        id: '123',
        name: 'Upgrade lodash to v4',
        description: `
# Removes lodash

- and renders in markdown
        `,
        changesets: { nodes: [{ externalState: GQL.ChangesetExternalState.OPEN }] },
        patches: { totalCount: 2 },
        createdAt: '2019-12-04T23:15:01Z',
        closedAt: null,
    }

    test('open campaign', () => {
        expect(
            mount(<CampaignNode node={node} now={parseISO('2019-01-01T23:15:01Z')} history={createMemoryHistory()} />)
        ).toMatchSnapshot()
    })
    test('closed campaign', () => {
        expect(
            mount(
                <CampaignNode
                    node={{ ...node, closedAt: '2019-12-04T23:19:01Z' }}
                    now={parseISO('2019-01-01T23:15:01Z')}
                    history={createMemoryHistory()}
                />
            )
        ).toMatchSnapshot()
    })
    test('campaign without description', () => {
        expect(
            mount(
                <CampaignNode
                    node={{
                        ...node,
                        // todo: make this null, once it's supported in the API: https://github.com/sourcegraph/sourcegraph/issues/9034
                        description: '',
                    }}
                    now={parseISO('2019-01-01T23:15:01Z')}
                    history={createMemoryHistory()}
                />
            )
        ).toMatchSnapshot()
    })
    test('campaign selection mode', () => {
        expect(
            mount(
                <CampaignNode
                    node={node}
                    selection={{
                        buttonLabel: 'Select',
                        enabled: true,
                        onSelect: () => undefined,
                    }}
                    now={parseISO('2019-01-01T23:15:01Z')}
                    history={createMemoryHistory()}
                />
            )
        ).toMatchSnapshot()
    })
    test('campaign with mixed changeset states', () => {
        expect(
            mount(
                <CampaignNode
                    node={{
                        ...node,
                        changesets: {
                            nodes: [
                                { externalState: GQL.ChangesetExternalState.OPEN },
                                { externalState: GQL.ChangesetExternalState.CLOSED },
                            ],
                        },
                    }}
                    now={parseISO('2019-01-01T23:15:01Z')}
                    history={createMemoryHistory()}
                />
            )
        ).toMatchSnapshot()
    })
})
