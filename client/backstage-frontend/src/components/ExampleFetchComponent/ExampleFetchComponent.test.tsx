import React from 'react'

import { setupRequestMockHandlers } from '@backstage/test-utils'
import { render, screen } from '@testing-library/react'
import { rest } from 'msw'
import { setupServer } from 'msw/node'

import { ExampleFetchComponent } from './ExampleFetchComponent'

describe('ExampleFetchComponent', () => {
    const server = setupServer()
    // Enable sane handlers for network requests
    setupRequestMockHandlers(server)

    // setup mock response
    beforeEach(() => {
        server.use(
            rest.get('https://randomuser.me/*', (_, res, ctx) => res(ctx.status(200), ctx.delay(2000), ctx.json({})))
        )
    })
    it('should render', async () => {
        await render(<ExampleFetchComponent />)
        expect(await screen.findByTestId('progress')).toBeInTheDocument()
    })
})
