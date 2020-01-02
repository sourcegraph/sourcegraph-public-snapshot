import React from 'react'
import sinon from 'sinon'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import renderer, { act } from 'react-test-renderer'
import { CampaignNamespaceField } from './CampaignNamespaceField'
import { NEVER, of } from 'rxjs'

const USER: Pick<GQL.Namespace, '__typename' | 'id' | 'namespaceName' | 'url'> = {
    __typename: 'User',
    id: 'u0',
    namespaceName: 'alice',
    url: 'https://example.com/u0',
}

describe('CampaignNamespaceField', () => {
    test('loading', () =>
        expect(
            renderer.create(
                <CampaignNamespaceField value={undefined} onChange={() => undefined} _queryNamespaces={() => NEVER} />
            )
        ).toMatchSnapshot())

    test('has initial value and calls onChange', () => {
        const onChange = sinon.spy()
        const component = renderer.create(
            <CampaignNamespaceField value={undefined} onChange={onChange} _queryNamespaces={() => of([USER])} />
        )
        act(() => undefined) // eslint-disable-line @typescript-eslint/no-floating-promises
        expect(component).toMatchSnapshot()

        expect(onChange.calledOnceWith(USER.id)).toBeTruthy()
    })
})
