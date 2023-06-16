/**
 * By default, Remix will handle hydrating your app on the client for you.
 * You are free to delete this file if you'd like to, but if you ever want it revealed again, you can run `npx remix reveal` ✨
 * For more information, see https://remix.run/file-conventions/entry.client
 */

import { startTransition, StrictMode } from 'react'

import { RemixBrowser } from '@remix-run/react'
import { hydrate } from 'react-dom'

// Throws on hydration 🫠
// import { hydrateRoot } from 'react-dom/client'

startTransition(() => {
    hydrate(
        <StrictMode>
            <RemixBrowser />
        </StrictMode>,
        document
    )
})
