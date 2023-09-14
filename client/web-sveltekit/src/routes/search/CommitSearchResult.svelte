<svelte:options immutable />

<script lang="ts" context="module">
    import hljs from 'highlight.js/lib/core'
    import diff from 'highlight.js/lib/languages/diff'
    import { highlightRanges } from '$lib/dom'

    hljs.registerLanguage('diff', diff)

    const highlightCommit: Action<HTMLElement, { ranges: [number, number][] }> = (node: HTMLElement, { ranges }) => {
        hljs.highlightElement(node)
        highlightRanges(node, { ranges })
    }

    function unwrapMarkdownCodeBlock(content: string): string {
        return content.replace(/^```[_a-z]*\n/i, '').replace(/\n```$/i, '')
    }

    function getMatches(result: CommitMatch): [number, number][] {
        const lines = unwrapMarkdownCodeBlock(result.content).split('\n')

        const lineOffsets: number[] = [0]
        for (let i = 1; i < lines.length; i++) {
            // Convert line to array of codepoints to get correct length
            lineOffsets[i] = lineOffsets[i - 1] + [...lines[i - 1]].length + 1
        }

        return result.ranges.map(([line, start, length]) => [
            lineOffsets[line - 1] + start,
            lineOffsets[line - 1] + start + length,
        ])
    }
</script>

<script lang="ts">
    import Timestamp from '$lib/Timestamp.svelte'
    import { displayRepoName, type CommitMatch, getRepositoryUrl, getMatchUrl } from '$lib/shared'
    import CodeHostIcon from './CodeHostIcon.svelte'
    import RepoStars from './RepoStars.svelte'
    import SearchResult from './SearchResult.svelte'
    import type { Action } from 'svelte/action'

    export let result: CommitMatch

    $: repoAtRevisionURL = getRepositoryUrl(result.repository)
    $: commitURL = getMatchUrl(result)
    $: subject = result.message.split('\n', 1)[0]
    $: commitOid = result.oid.slice(0, 7)
    $: content = unwrapMarkdownCodeBlock(result.content)
    $: matches = getMatches(result)
    let highlightCls: string
    $: {
        const lang = /```(\S+)\s/.exec(result.content)?.[1]
        // highlight.js logs a warning if the defined language isn't
        // registers, which is noisy in the console and in tests
        highlightCls = lang?.toLowerCase() === 'diff' ? `language-${lang}` : 'no-highlight'
    }
</script>

<SearchResult>
    <CodeHostIcon slot="icon" repository={result.repository} />
    <div slot="title" data-sveltekit-preload-data="tap">
        <a href={repoAtRevisionURL}>{displayRepoName(result.repository)}</a>
        <span aria-hidden={true}>›</span>
        <a href={commitURL}>{result.authorName}</a>
        <span aria-hidden={true}>:&nbsp;</span>
        <a href={commitURL}>{subject}</a>
    </div>
    <svelte:fragment slot="info">
        <a href={commitURL} data-sveltekit-preload-data="tap">
            <code>{commitOid}</code>
            &nbsp;
            <Timestamp date={result.committerDate} strict utc />
        </a>
        {#if result.repoStars}
            <span class="divider" />
            <RepoStars repoStars={result.repoStars} />
        {/if}
    </svelte:fragment>
    <!-- #key is needed here to recreate the element because use:highlightCommit changes the DOM -->
    {#key content}
        <pre class="{highlightCls} p-2" use:highlightCommit={{ ranges: matches }}>{content}</pre>
    {/key}
</SearchResult>

<style lang="scss">
    .divider {
        border-left: 1px solid var(--border-color);
        padding-left: 0.5rem;
        margin-left: 0.5rem;
    }

    code {
        background: var(--code-bg);
        display: inline-block;
        padding: 0.25rem;
    }

    pre {
        margin: 0;
        font-family: var(--code-font-family);
        font-size: var(--code-font-size);
    }
</style>
