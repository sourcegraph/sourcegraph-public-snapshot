import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignNode } from './CampaignNode'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { parseISO } from 'date-fns'
import { createMemoryHistory } from 'history'

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
        changesets: { nodes: [{ state: GQL.ChangesetState.OPEN }] },
        patches: { totalCount: 2 },
        createdAt: '2019-12-04T23:15:01Z',
        closedAt: null,
    }

    test('open campaign', () => {
        expect(
            renderer
                .create(
                    <CampaignNode node={node} now={parseISO('2019-01-01T23:15:01Z')} history={createMemoryHistory()} />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
    test('closed campaign', () => {
        expect(
            renderer
                .create(
                    <CampaignNode
                        node={{ ...node, closedAt: '2019-12-04T23:19:01Z' }}
                        now={parseISO('2019-01-01T23:15:01Z')}
                        history={createMemoryHistory()}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
    test('campaign without description', () => {
        expect(
            renderer
                .create(
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
                .toJSON()
        ).toMatchSnapshot()
    })
    test('campaign selection mode', () => {
        expect(
            renderer
                .create(
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
                .toJSON()
        ).toMatchSnapshot()
    })
    test('campaign with mixed changeset states', () => {
        expect(
            renderer
                .create(
                    <CampaignNode
                        node={{
                            ...node,
                            changesets: {
                                nodes: [{ state: GQL.ChangesetState.OPEN }, { state: GQL.ChangesetState.CLOSED }],
                            },
                        }}
                        now={parseISO('2019-01-01T23:15:01Z')}
                        history={createMemoryHistory()}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
