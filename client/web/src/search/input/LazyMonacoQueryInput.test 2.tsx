import { noop } from 'lodash'
import React from 'react'
import renderer from 'react-test-renderer'

import { PlainQueryInput } from './LazyMonacoQueryInput'

describe('PlainQueryInput', () => {
    test('empty', () =>
        expect(
            renderer
                .create(
                    <PlainQueryInput
                        queryState={{
                            query: '',
                        }}
                        onChange={noop}
                    />
                )
                .toJSON()
        ).toMatchSnapshot())

    test('with query', () =>
        expect(
            renderer
                .create(
                    <PlainQueryInput
                        queryState={{
                            query: 'repo:jsonrpc2 file:async.go asyncHandler',
                        }}
                        onChange={noop}
                    />
                )
                .toJSON()
        ).toMatchSnapshot())
})
