<script lang="ts">
    import { page } from '$app/stores'
    import { SourcegraphURL } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import FileDiffHunks from '$lib/repo/FileDiffHunks.svelte'
    import FileHeader from '$lib/repo/FileHeader.svelte'
    import FileIcon from '$lib/repo/FileIcon.svelte'
    import { toReadable } from '$lib/utils'
    import { Alert } from '$lib/wildcard'

    import type { PageData } from './$types'
    import DiffSummaryHeader from './DiffSummaryHeader.svelte'

    export let data: Extract<PageData, { type: 'DiffView' }>

    $: commit = toReadable(data.commit)
    $: hunks = $commit.value?.diff.fileDiffs.nodes.at(0)?.hunks
</script>

<FileHeader type="blob" repoName={data.repoName} revision={data.revision} path={data.filePath}>
    <svelte:fragment slot="file-icon">
        {#if $commit.value?.blob}
            <FileIcon file={$commit.value.blob} inline />
        {/if}
    </svelte:fragment>
</FileHeader>

<div class="info">
    {#if $commit.value}
        <DiffSummaryHeader commit={$commit.value} />
        <a href={SourcegraphURL.from($page.url).deleteSearchParameter('rev', 'diff').toString()}>
            <Icon icon={ILucideX} inline aria-hidden />
            <span>Close diff</span>
        </a>
    {/if}
</div>

<div class="content" class:center={!!($commit.error || !hunks)}>
    {#if $commit.pending}
        <LoadingSpinner />
    {:else if $commit.error}
        <Alert variant="danger">Unable to load diff information: {$commit.error.message}</Alert>
    {:else if !hunks}
        <Alert variant="danger">Unable to load diff information.</Alert>
    {:else}
        <FileDiffHunks {hunks} />
    {/if}
</div>

<style lang="scss">
    .content {
        overflow: auto;
        flex: 1;

        &.center {
            align-items: center;
            display: flex;
            flex-direction: column;
        }
    }

    .info {
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 1rem;
        overflow: hidden;

        padding: 0.5rem 1rem;
        color: var(--text-muted);
        background-color: var(--code-bg);
        box-shadow: var(--file-header-shadow);

        // Allows for its shadow to cascade over the code panel
        z-index: 1;
        border-top: 1px solid var(--border-color);

        @media (--mobile) {
            padding: 0.25rem 0.5rem;
        }

        a {
            text-decoration: none;
            white-space: nowrap;

            --icon-color: var(--link-color);

            // This is used to avoid having the whitespace being underlined on hover
            &:hover span {
                text-decoration: underline;
            }

            @media (--mobile) {
                // Hide link label on mobile
                span {
                    display: none;
                }
            }
        }
    }
</style>
