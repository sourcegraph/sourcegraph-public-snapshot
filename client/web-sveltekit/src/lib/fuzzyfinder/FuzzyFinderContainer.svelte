<script lang="ts" context="module">
    import { writable } from 'svelte/store'
    import type { FuzzyFinderTabId } from './FuzzyFinder.svelte'

    interface FuzzyFinderState {
        open: boolean
        selectedTabId: FuzzyFinderTabId | ''
    }

    const fuzzyFinderState = writable<FuzzyFinderState>({ open: false, selectedTabId: '' })

    export function openFuzzyFinder(tab?: FuzzyFinderTabId): void {
        fuzzyFinderState.update(state => ({ selectedTabId: tab ?? state.selectedTabId, open: true }))
    }
</script>

<script lang="ts">
    import { escapeRegExp } from 'lodash'

    import { page } from '$app/stores'
    import { registerHotkey } from '$lib/Hotkey'
    import { parseRepoRevision } from '$lib/shared'

    import FuzzyFinder from './FuzzyFinder.svelte'
    import { allHotkey, filesHotkey, reposHotkey, symbolsHotkey } from './keys'

    let finder: FuzzyFinder | undefined
    let scope = ''

    registerHotkey({
        keys: allHotkey,
        ignoreInputFields: false,
        handler: event => {
            event.stopPropagation()
            fuzzyFinderState.set({
                open: true,
                selectedTabId: 'all',
            })
            return false
        },
    })

    registerHotkey({
        keys: reposHotkey,
        ignoreInputFields: false,
        handler: event => {
            event.stopPropagation()
            fuzzyFinderState.set({
                open: true,
                selectedTabId: 'repos',
            })
            return false
        },
    })

    registerHotkey({
        keys: symbolsHotkey,
        ignoreInputFields: false,
        handler: event => {
            event.stopPropagation()
            fuzzyFinderState.set({
                open: true,
                selectedTabId: 'symbols',
            })
            return false
        },
    })

    registerHotkey({
        keys: filesHotkey,
        ignoreInputFields: false,
        handler: event => {
            event.stopPropagation()
            fuzzyFinderState.set({
                open: true,
                selectedTabId: 'files',
            })
            return false
        },
    })

    registerHotkey({
        keys: { key: 'Esc' },
        ignoreInputFields: false,
        handler: event => {
            event.preventDefault()
            fuzzyFinderState.update(state => ({ ...state, open: false }))
            return false
        },
    })

    $: if ($fuzzyFinderState.selectedTabId !== '') {
        finder?.selectTab($fuzzyFinderState.selectedTabId)
    }

    $: if ($page.params.repo) {
        const { repoName, revision } = parseRepoRevision($page.params.repo)
        scope = `repo:^${escapeRegExp(repoName)}$`
        if (revision) {
            scope += `@${revision}`
        }
    } else {
        scope = ''
    }
</script>

<FuzzyFinder
    bind:this={finder}
    {scope}
    open={$fuzzyFinderState.open}
    on:close={() => ($fuzzyFinderState.open = false)}
/>
