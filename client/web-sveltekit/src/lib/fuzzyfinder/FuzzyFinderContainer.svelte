<script lang="ts">
    import { escapeRegExp } from 'lodash'

    import { page } from '$app/stores'
    import { registerHotkey } from '$lib/Hotkey'
    import { parseRepoRevision } from '$lib/shared'

    import FuzzyFinder from './FuzzyFinder.svelte'
    import { filesHotkey, reposHotkey, symbolsHotkey } from './keys'

    let open = false
    let finder: FuzzyFinder | undefined
    let scope = ''

    registerHotkey({
        keys: reposHotkey,
        ignoreInputFields: false,
        handler: () => {
            open = true
            finder?.selectTab('repos')
        },
    })
    registerHotkey({
        keys: symbolsHotkey,
        ignoreInputFields: false,
        handler: () => {
            open = true
            finder?.selectTab('symbols')
        },
    })

    registerHotkey({
        keys: filesHotkey,
        ignoreInputFields: false,
        handler: () => {
            open = true
            finder?.selectTab('files')
        },
    })

    $: if ($page.params.repo) {
        const { repoName } = parseRepoRevision($page.params.repo)
        scope = `repo:^${escapeRegExp(repoName)}$`
    } else {
        scope = ''
    }
</script>

<FuzzyFinder bind:this={finder} {open} {scope} on:close={() => (open = false)} />
