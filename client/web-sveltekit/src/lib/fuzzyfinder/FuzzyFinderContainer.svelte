<script lang="ts" context="module">
    import { writable } from 'svelte/store'

    const fuzzyFinderState = writable(false)
    const scopeState = writable<FuzzyFinderTabId | ''>('')

    export function openFuzzyFinder(tab?: FuzzyFinderTabId): void {
        fuzzyFinderState.set(true)
        scopeState.set(tab ?? FuzzyFinderTabType.Repos)
    }
</script>

<script lang="ts">
    import { escapeRegExp } from 'lodash'

    import { page } from '$app/stores'
    import { registerHotkey } from '$lib/Hotkey'
    import { parseRepoRevision } from '$lib/shared'

    import FuzzyFinder, { type FuzzyFinderTabId, FuzzyFinderTabType } from './FuzzyFinder.svelte'
    import { filesHotkey, reposHotkey, symbolsHotkey } from './keys'

    let finder: FuzzyFinder | undefined
    let scope = ''

    registerHotkey({
        keys: reposHotkey,
        ignoreInputFields: false,
        handler: event => {
            event.stopPropagation()
            $fuzzyFinderState = true
            $scopeState = 'repos'
            return false
        },
    })
    registerHotkey({
        keys: symbolsHotkey,
        ignoreInputFields: false,
        handler: event => {
            event.stopPropagation()
            $fuzzyFinderState = true
            $scopeState = 'symbols'
            return false
        },
    })

    registerHotkey({
        keys: filesHotkey,
        ignoreInputFields: false,
        handler: event => {
            event.stopPropagation()
            $fuzzyFinderState = true
            $scopeState = 'files'
            return false
        },
    })

    scopeState.subscribe(tab => {
        if (tab) {
            finder?.selectTab(tab)
        }
    })

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

<FuzzyFinder bind:this={finder} open={$fuzzyFinderState} {scope} on:close={() => ($fuzzyFinderState = false)} />
