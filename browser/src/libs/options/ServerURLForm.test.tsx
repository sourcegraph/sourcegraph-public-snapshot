import renderer from 'react-test-renderer'
import * as React from 'react'

import { ServerURLForm, ServerURLFormProps } from './ServerURLForm'
import { Omit } from 'utility-types'
import { InvalidSourcegraphURLError } from '../../shared/util/context'
import { AuthRequiredError } from '../../../../shared/src/backend/errors'

describe('ServerURLForm', () => {
    const props: Omit<ServerURLFormProps, 'sourcegraphURL' | 'connectionStatus'> = {
        onSourcegraphURLChange: jest.fn(),
        onSourcegraphURLSubmit: jest.fn(),
        requestSourcegraphURLPermissions: jest.fn(),
    }
    test('Editing', () => {
        expect(
            renderer
                .create(<ServerURLForm {...props} sourcegraphURL="https://sourcegraph" connectionStatus={undefined} />)
                .toJSON()
        ).toMatchSnapshot()
    })

    test('Connecting', () => {
        expect(
            renderer
                .create(
                    <ServerURLForm
                        {...props}
                        sourcegraphURL="https://sourcegraph.com"
                        connectionStatus={{ type: 'connecting' }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('Connected', () => {
        expect(
            renderer
                .create(
                    <ServerURLForm
                        {...props}
                        sourcegraphURL="https://sourcegraph.com"
                        connectionStatus={{ type: 'connected' }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('Invalid URL', () => {
        expect(
            renderer
                .create(
                    <ServerURLForm
                        {...props}
                        sourcegraphURL="sourcegraph.com"
                        connectionStatus={{ type: 'error', error: new InvalidSourcegraphURLError('sourcegraph.com') }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('Auth required', () => {
        expect(
            renderer
                .create(
                    <ServerURLForm
                        {...props}
                        sourcegraphURL="https://sourcegraph.sgdev.org"
                        connectionStatus={{
                            type: 'error',
                            error: new AuthRequiredError('https://sourcegraph.sgdev.org'),
                        }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('Other error', () => {
        expect(
            renderer
                .create(
                    <ServerURLForm
                        {...props}
                        sourcegraphURL="https://sourcegraph.sgdev.org"
                        connectionStatus={{
                            type: 'error',
                            error: new Error(),
                        }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('Other error - has permissions', () => {
        expect(
            renderer
                .create(
                    <ServerURLForm
                        {...props}
                        sourcegraphURL="https://sourcegraph.sgdev.org"
                        connectionStatus={{
                            type: 'error',
                            error: new Error(),
                            urlHasPermissions: true,
                        }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('Other error - no permissions', () => {
        expect(
            renderer
                .create(
                    <ServerURLForm
                        {...props}
                        sourcegraphURL="https://sourcegraph.sgdev.org"
                        connectionStatus={{
                            type: 'error',
                            error: new Error(),
                            urlHasPermissions: false,
                        }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
