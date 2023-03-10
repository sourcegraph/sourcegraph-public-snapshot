import React from 'react'

import { setupRequestMockHandlers, renderInTestApp } from '@backstage/test-utils'
import { screen } from '@testing-library/react'
import { rest } from 'msw'
import { setupServer } from 'msw/node'

import { ExampleComponent } from './ExampleComponent'

describe('ExampleComponent', () => {
    const server = setupServer()
    // Enable sane handlers for network requests
    setupRequestMockHandlers(server)

    // setup mock response
    beforeEach(() => {
        server.use(rest.get('/*', (_, res, ctx) => res(ctx.status(200), ctx.json({}))))
    })

    it('should render', async () => {
        await renderInTestApp(<ExampleComponent />)
        expect(screen.getByText('Welcome to Sourcgraph!')).toBeInTheDocument()
    })
})
