<script context="module" lang="ts">
</script>

<script lang="ts">
    import { ExternalServiceKind } from '$lib/graphql-types'
    import DisplayPath from '$lib/path/DisplayPath.svelte'
    import Tooltip from '$lib/Tooltip.svelte'

    import { inferSplitCodeHost } from './codehost'
    import CodeHostIcon from './codehost/CodeHostIcon.svelte'

    export let repoName: string
    export let kind: ExternalServiceKind | undefined // No default to discourage inferring kind

    $: ({ kind: inferredKind, codeHost, displayName } = inferSplitCodeHost(repoName, kind))
</script>

<!--
    NOTE: the awkward formatting is intentional to ensure there is no extra
    whitespace in the DOM. This is necessary for correctly highlighting
    matches inside a repo name by offset.
-->
<span
    ><!--
    --><Tooltip tooltip={codeHost}
        ><!--
        --><CodeHostIcon kind={inferredKind} inline /><!--
    --></Tooltip
    ><!--
    --><span
        ><!-- Include the tooltip text hidden in the DOM for the purpose of highlighting
        --><span class="hidden"
            >{codeHost ? codeHost + '/' : ''}</span
        ><!--
        --><DisplayPath path={displayName} /><!--
    --></span
    ><!--
--></span
>

<style lang="scss">
    span {
        display: inline-flex;
        align-items: baseline;
        gap: 0.5em;

        :global([data-icon]) {
            align-self: center;
        }

        :global([data-path-container]) {
            font-family: var(--font-family-base);
        }

        .hidden {
            display: none;
        }
    }
</style>
