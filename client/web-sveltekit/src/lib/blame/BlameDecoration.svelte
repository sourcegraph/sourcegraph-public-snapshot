<script lang="ts">
    import type { BlameHunk } from '$lib/web'
    import { formatDate, TimestampFormat } from '$lib/Timestamp.svelte'

    export let hunk: BlameHunk

    // These are in the React implementation, but not yet used in the Svlete implementation.
    // Leaving them commented for reference if/when we implement the popover.
    //
    // export let line: number
    // export let onSelect: (line: number) => void
    // export let onDeselect: (line: number) => void
    // export let externalURLs: BlameHunkData['externalURLs']

    $: info = hunk.displayInfo
</script>

<div class="root">
    <span class="date">{formatDate(info.commitDate, { format: TimestampFormat.FULL_DATE })}</span>
    <a href={info.linkURL} target="_blank" rel="noreferrer noopener">
        <span class="name">{`${info.displayName}${info.username}`.split(' ')[0]}</span>
        <span class="message">{info.message}</span>
    </a>
</div>

<style lang="scss">
    .root {
        font-family: var(--font-family-base);
        color: var(--text-muted);
        display: flex;
        align-items: center;
        padding: 0 0.5em;
        gap: 0.5em;

        .date {
            flex-shrink: 0;
            font-weight: bold;
            margin-right: 0.25rem;
        }

        a {
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            color: inherit;
            width: 100%;
        }

        .date,
        .name {
            font-family: monospace;
        }

        .message {
            margin-left: 0.5rem;
        }
    }
</style>
