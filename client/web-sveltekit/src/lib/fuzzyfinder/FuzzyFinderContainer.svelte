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
        handler: event => {
            event.stopPropagation()
            open = true
            finder?.selectTab('repos')
            return false
        },
    })
    registerHotkey({
        keys: symbolsHotkey,
        ignoreInputFields: false,
        handler: event => {
            event.stopPropagation()
            open = true
            finder?.selectTab('symbols')
            return false
        },
    })

    registerHotkey({
        keys: filesHotkey,
        ignoreInputFields: false,
        handler: event => {
            event.stopPropagation()
            open = true
            finder?.selectTab('files')
            return false
        },
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

<FuzzyFinder bind:this={finder} {open} {scope} on:close={() => (open = false)} />
