<script lang="ts">
    import { Meta, Story, Template } from '@storybook/addon-svelte-csf'
    import SearchResults, { setSearchResultsContext } from './SearchResults.svelte'
    import {
        createCommitMatch,
        createContentMatch,
        createHighlightedFileResult,
        createPathMatch,
        createPersonMatch,
        createSymbolMatch,
        createTeamMatch,
    } from '$testdata'
    import FileContentSearchResult from './FileContentSearchResult.svelte'
    import { SvelteComponent, setContext } from 'svelte'
    import { KEY, type SourcegraphContext } from '$lib/stores'
    import { readable } from 'svelte/store'
    import CommitSearchResult from './CommitSearchResult.svelte'
    import PersonSearchResult from './PersonSearchResult.svelte'
    import TeamSearchResult from './TeamSearchResult.svelte'
    import { queryStateStore } from '$lib/search/state'
    import { graphql } from 'msw'
    import type { HighlightedFileResult, HighlightedFileVariables } from '$lib/graphql-operations'
    import type { SearchMatch } from '$lib/shared'
    import FilePathSearchResult from './FilePathSearchResult.svelte'
    import SymbolSearchResult from './SymbolSearchResult.svelte'
    import { createTemporarySettingsStorage } from '$lib/temporarySettings'

    setContext<SourcegraphContext>(KEY, {
        user: readable(null),
        settings: readable({}),
        isLightTheme: readable(true),
        featureFlags: readable([]),
        temporarySettingsStorage: createTemporarySettingsStorage(),
        client: readable(null),
    })

    setSearchResultsContext({
        isExpanded(_match) {
            return false
        },
        setExpanded(_match, _expanded) {},
        queryState: queryStateStore(undefined, {}),
    })
    // TS complains about up MockSuitFunctions which is not relevant here
    // @ts-ignore
    window.context = { xhrHeaders: {} }

    const results: [string, typeof SvelteComponent<{ result: SearchMatch }>, () => SearchMatch][] = [
        ['Path match', FilePathSearchResult, createPathMatch],
        ['Content match', FileContentSearchResult, createContentMatch],
        ['Commit match', CommitSearchResult, () => createCommitMatch('commit')],
        ['Commit match (diff)', CommitSearchResult, () => createCommitMatch('diff')],
        ['Symbol match', SymbolSearchResult, createSymbolMatch],
        ['Person match', PersonSearchResult, createPersonMatch],
        ['Team match', TeamSearchResult, createTeamMatch],
    ]

    const data = results.map(([, , generator]) => generator())

    function randomizeData(i: number) {
        data[i] = results[i][2]()
    }

    $: parameters = {
        msw: {
            handlers: {
                highlightedFile: graphql.query<HighlightedFileResult, HighlightedFileVariables>(
                    'HighlightedFile',
                    (req, res, ctx) => res(ctx.data(createHighlightedFileResult(req.variables.ranges)))
                ),
            },
        },
    }
</script>

<Meta title="search/SearchResults" component={SearchResults} {parameters} />

<Template>
    {#each results as [title, component], i}
        <div>
            <h2>{title}</h2>
            <button on:click={() => randomizeData(i)}>Randomize</button>
        </div>
        <svelte:component this={component} result={data[i]} />
    {/each}
</Template>

<Story name="Default" />

<style lang="scss">
    div {
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    h2 {
        margin: 1rem 0;
    }
</style>
