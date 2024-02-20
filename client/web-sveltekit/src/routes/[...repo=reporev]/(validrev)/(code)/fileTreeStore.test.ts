import { basename } from 'path'

import { test, expect, vi, beforeEach, afterEach } from 'vitest'

import { createFileTreeStore, clearTopTreePathCache_testingOnly } from './fileTreeStore'

vi.mock('$app/environment', () => ({ browser: true }))

let store: ReturnType<typeof createFileTreeStore>
let fetchFileTreeData: Parameters<typeof createFileTreeStore>[0]['fetchFileTreeData'] = async ({ filePath }) => ({
    // The exact values don't matter for this test
    root: {
        name: basename(filePath),
        path: filePath,
        isRoot: filePath === '',
        entries: [],
        isDirectory: true,
        canonicalURL: '',
    },
    values: [],
})

beforeEach(async () => {
    vi.useFakeTimers()
    store = createFileTreeStore({
        fetchFileTreeData,
    })
})

afterEach(async () => {
    vi.useRealTimers()
    clearTopTreePathCache_testingOnly()
})

test('caches top-level tree paths in the browser', async () => {
    const subscriber = vi.fn()
    store.subscribe(subscriber)

    // Start a specific file tree
    store.set({ repoName: 'repo', revision: 'rev', path: 'a/b/c/d' })
    await vi.runAllTimersAsync()

    expect(subscriber.mock.lastCall?.[0]?.getRoot()).toMatchObject({ path: 'a/b/c/d' })

    // Navigate up the tree -> a new provider should be created
    store.set({ repoName: 'repo', revision: 'rev', path: 'a/b' })
    await vi.runAllTimersAsync()

    expect(subscriber.mock.lastCall?.[0]?.getRoot()).toMatchObject({ path: 'a/b' })

    // Navigate down the tree -> the previous provider should be reused
    store.set({ repoName: 'repo', revision: 'rev', path: 'a/b/c' })
    await vi.runAllTimersAsync()

    expect(subscriber.mock.lastCall?.[0]?.getRoot()).toMatchObject({ path: 'a/b' })
})

test('per repository caching', async () => {
    const subscriber = vi.fn()
    store.subscribe(subscriber)

    // Start a specific file tree
    store.set({ repoName: 'repo', revision: 'rev', path: 'a/b/c' })
    await vi.runAllTimersAsync()

    expect(subscriber.mock.lastCall?.[0]?.getRoot()).toMatchObject({ path: 'a/b/c' })

    // Navigate to different repo -> a new provider should be created
    store.set({ repoName: 'repo2', revision: 'rev', path: 'a/b/c/d/e' })
    await vi.runAllTimersAsync()

    expect(subscriber.mock.lastCall?.[0]?.getRoot()).toMatchObject({ path: 'a/b/c/d/e' })

    // Navigate back to the first repo -> the previous provider should be reused
    store.set({ repoName: 'repo', revision: 'rev', path: 'a/b/c/d' })
    await vi.runAllTimersAsync()

    expect(subscriber.mock.lastCall?.[0]?.getRoot()).toMatchObject({ path: 'a/b/c' })
})

test('per revision caching', async () => {
    const subscriber = vi.fn()
    store.subscribe(subscriber)

    // Start a specific file tree
    store.set({ repoName: 'repo', revision: 'rev', path: 'a/b/c' })
    await vi.runAllTimersAsync()

    expect(subscriber.mock.lastCall?.[0]?.getRoot()).toMatchObject({ path: 'a/b/c' })

    // Navigate to different revision -> a new provider should be created
    store.set({ repoName: 'repo', revision: 'rev1', path: 'a/b/c/d/e' })
    await vi.runAllTimersAsync()

    expect(subscriber.mock.lastCall?.[0]?.getRoot()).toMatchObject({ path: 'a/b/c/d/e' })

    // Navigate back to the prevision revision -> the previous provider should be reused
    store.set({ repoName: 'repo', revision: 'rev', path: 'a/b/c/d' })
    await vi.runAllTimersAsync()

    expect(subscriber.mock.lastCall?.[0]?.getRoot()).toMatchObject({ path: 'a/b/c' })
})

test('error recovery', async () => {
    function getStore(): ReturnType<typeof createFileTreeStore> {
        return createFileTreeStore({
            fetchFileTreeData: vi.fn(fetchFileTreeData).mockRejectedValueOnce(new Error('Error fetching file tree')),
        })
    }
    let store = getStore()

    const subscriber = vi.fn()
    store.subscribe(subscriber)

    // Start a specific file tree
    store.set({ repoName: 'repo', revision: 'rev', path: 'a/b/c' })
    await vi.runAllTimersAsync()
    expect(subscriber.mock.lastCall?.[0]).toBeInstanceOf(Error)

    // Navigate up the tree
    store.set({ repoName: 'repo', revision: 'rev', path: 'a/b' })
    await vi.runAllTimersAsync()
    expect(subscriber.mock.lastCall?.[0]).not.toBeInstanceOf(Error)
    expect(subscriber.mock.lastCall?.[0]?.getRoot()).toMatchObject({ path: 'a/b' })

    // Reset to test error recovery when switching to a different path
    store = getStore()
    subscriber.mockClear()
    store.subscribe(subscriber)

    // Start a specific file tree
    store.set({ repoName: 'repo', revision: 'rev', path: 'a/b/c' })
    await vi.runAllTimersAsync()
    expect(subscriber.mock.lastCall?.[0]).toBeInstanceOf(Error)

    // Navigate to other path
    store.set({ repoName: 'repo', revision: 'rev', path: 'a/x/y' })
    await vi.runAllTimersAsync()
    expect(subscriber.mock.lastCall?.[0]).not.toBeInstanceOf(Error)
    expect(subscriber.mock.lastCall?.[0]?.getRoot()).toMatchObject({ path: 'a/x/y' })
})
