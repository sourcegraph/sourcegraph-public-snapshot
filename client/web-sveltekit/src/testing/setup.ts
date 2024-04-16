import { faker } from '@faker-js/faker'
import { cleanup } from '@testing-library/svelte'
import { vi, beforeEach, afterEach, beforeAll, afterAll } from 'vitest'

import { mockedContexts } from '$testing/mocks'

type SvelteAPI = typeof import('svelte')

// Mock Svelte's context API. This API is usually only available
// when rendering components.
vi.mock('svelte', async (actual: () => Promise<SvelteAPI>): Promise<SvelteAPI> => {
    const svelte = await actual()
    return {
        ...svelte,
        setContext(key, value) {
            if (mockedContexts.has(key)) {
                mockedContexts.set(key, value)
            } else {
                svelte.setContext(key, value)
            }
            return value
        },
        getContext(key) {
            if (mockedContexts.has(key)) {
                return mockedContexts.get(key)
            }
            try {
                return svelte.getContext(key)
            } catch {
                throw new Error(`Unable to get context '${key}'. Maybe you want to mock it?`)
            }
        },
    }
})

beforeAll(() => {
    // window.context is accessed by some existing modules
    vi.stubGlobal('context', {})
})

beforeEach(() => {
    // Set fixed date and faker seed for each tests
    const date = new Date(2021, 4, 24, 12, 0, 0)
    faker.setDefaultRefDate(date)
    faker.seed(24)
})

afterEach(() => {
    faker.setDefaultRefDate()
    faker.seed()

    // We need to call this manually because we don't use global
    // test functions.
    cleanup()
})

afterAll(() => {
    vi.unstubAllGlobals()
})
