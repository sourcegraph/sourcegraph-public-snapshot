<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import DisplayPath from '$lib/path/DisplayPath.svelte'
    import Tooltip from '$lib/Tooltip.svelte'

    import { type CodeHostKind, getHumanNameForCodeHost, getIconForCodeHost, inferCodeHost } './codehost'
    import { displayRepoName } from './index'

    export let repoName: string
    export let codeHost: CodeHostKind | undefined

    const {name: codeHostName, kind: codeHostKind} =
        codeHost
            ? {kind: codeHost, name: getHumanNameForCodeHost(codeHost)}
            : inferCodeHost(repoName)

    $: displayName = displayRepoName(repoName)
</script>

<Tooltip tooltip={codeHostName}>
    <Icon icon={getIconForCodeHost(codeHostKind)} inline aria-label="" />
</Tooltip>
<DisplayPath path={displayName} />

<style lang="scss">
    span {
        display: contents;
        :global([data-path-container]) {
            font-family: var(--font-family-base);
            font-size: var(--font-size-base);
        }
    }
</style>
