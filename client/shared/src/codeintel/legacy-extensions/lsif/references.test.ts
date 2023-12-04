import * as assert from 'assert'

import * as sinon from 'sinon'
import { describe, it } from 'vitest'

import * as sourcegraph from '../api'
import type { QueryGraphQLFn } from '../util/graphql'

import type { GenericLSIFResponse } from './api'
import { type ReferencesResponse, MAX_REFERENCE_PAGE_REQUESTS, referencesForPosition } from './references'
import {
    gatherValues,
    makeEnvelope,
    resource1,
    resource2,
    resource3,
    range1,
    range2,
    range3,
    document,
    position,
} from './util.test'

describe('referencesForPosition', () => {
    it('should correctly parse result', async () => {
        const queryGraphQLFn = sinon.spy<QueryGraphQLFn<GenericLSIFResponse<ReferencesResponse | null>>>(() =>
            makeEnvelope({
                references: {
                    nodes: [
                        { resource: resource1, range: range1 },
                        { resource: resource2, range: range2 },
                        { resource: resource3, range: range3 },
                    ],
                    pageInfo: {},
                },
            })
        )

        assert.deepEqual(await gatherValues(referencesForPosition(document, position, queryGraphQLFn)), [
            [
                new sourcegraph.Location(new URL('git://repo1?deadbeef1#a.ts'), range1),
                new sourcegraph.Location(new URL('git://repo2?deadbeef2#b.ts'), range2),
                new sourcegraph.Location(new URL('git://repo3?deadbeef3#c.ts'), range3),
            ],
        ])
    })

    it('should deal with empty payload', async () => {
        const queryGraphQLFn = sinon.spy<QueryGraphQLFn<GenericLSIFResponse<ReferencesResponse | null>>>(() =>
            makeEnvelope()
        )

        assert.deepEqual(await gatherValues(referencesForPosition(document, position, queryGraphQLFn)), [])
    })

    it('should paginate results', async () => {
        const stub = sinon.stub<
            Parameters<QueryGraphQLFn<GenericLSIFResponse<ReferencesResponse | null>>>,
            ReturnType<QueryGraphQLFn<GenericLSIFResponse<ReferencesResponse | null>>>
        >()
        const queryGraphQLFn = sinon.spy<QueryGraphQLFn<GenericLSIFResponse<ReferencesResponse | null>>>(stub)

        stub.onCall(0).returns(
            makeEnvelope({
                references: {
                    nodes: [{ resource: resource1, range: range1 }],
                    pageInfo: { endCursor: 'page2' },
                },
            })
        )
        stub.onCall(1).returns(
            makeEnvelope({
                references: {
                    nodes: [{ resource: resource2, range: range2 }],
                    pageInfo: { endCursor: 'page3' },
                },
            })
        )
        stub.onCall(2).returns(
            makeEnvelope({
                references: {
                    nodes: [{ resource: resource3, range: range3 }],
                    pageInfo: {},
                },
            })
        )

        const location1 = new sourcegraph.Location(new URL('git://repo1?deadbeef1#a.ts'), range1)
        const location2 = new sourcegraph.Location(new URL('git://repo2?deadbeef2#b.ts'), range2)
        const location3 = new sourcegraph.Location(new URL('git://repo3?deadbeef3#c.ts'), range3)

        assert.deepEqual(await gatherValues(referencesForPosition(document, position, queryGraphQLFn)), [
            [location1],
            [location1, location2],
            [location1, location2, location3],
        ])

        assert.strictEqual(queryGraphQLFn.getCall(0).args[1]?.after, undefined)
        assert.strictEqual(queryGraphQLFn.getCall(1).args[1]?.after, 'page2')
        assert.strictEqual(queryGraphQLFn.getCall(2).args[1]?.after, 'page3')
    })

    it('should not page results indefinitely', async () => {
        const queryGraphQLFn = sinon.spy<QueryGraphQLFn<GenericLSIFResponse<ReferencesResponse | null>>>(() =>
            makeEnvelope({
                references: {
                    nodes: [{ resource: resource1, range: range1 }],
                    pageInfo: { endCursor: 'infinity' },
                },
            })
        )

        const location = new sourcegraph.Location(new URL('git://repo1?deadbeef1#a.ts'), range1)

        const values = [[location]]
        for (let index = 1; index < MAX_REFERENCE_PAGE_REQUESTS; index++) {
            const lastCopy = [...values.at(-1)!]
            lastCopy.push(location)
            values.push(lastCopy)
        }

        assert.deepEqual(await gatherValues(referencesForPosition(document, position, queryGraphQLFn)), values)

        assert.strictEqual(queryGraphQLFn.callCount, MAX_REFERENCE_PAGE_REQUESTS)
    })
})
