<script lang="ts">
    import copy from 'copy-to-clipboard'

    import Icon2 from '$lib/Icon2.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Button } from '$lib/wildcard'

    export let value: string
    export let label = 'Copy to clipboard'

    let recentlyCopied = false
    function handleCopyPath(): void {
        copy(value)
        recentlyCopied = true
        setTimeout(() => {
            recentlyCopied = false
        }, 1000)
    }

    $: tooltip = recentlyCopied ? 'Copied!' : label
</script>

<span class="copy-path-button">
    <Tooltip {tooltip} placement="bottom">
        <Button on:click={handleCopyPath} variant="icon" size="sm" aria-label={label}>
            <Icon2 inline icon={ILucideCopy} aria-hidden />
        </Button>
    </Tooltip>
</span>

<style lang="scss">
    .copy-path-button {
        display: contents;

        --color: var(--icon-color);
        &:hover {
            --color: var(--body-color);
        }
    }
</style>
