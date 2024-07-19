<script lang="ts">
    import copy from 'copy-to-clipboard'

    import Icon from '$lib/Icon.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Button } from '$lib/wildcard'

    export let value: string
    export let label = 'Copy to clipboard'

    let recentlyCopied = false
    function handleCopy(): void {
        copy(value)
        recentlyCopied = true
        setTimeout(() => {
            recentlyCopied = false
        }, 1000)
    }

    $: tooltip = recentlyCopied ? 'Copied!' : label
</script>

<span class="copy-button">
    <Tooltip {tooltip} placement="bottom"
        ><slot name="custom" {handleCopy}
            ><Button on:click={handleCopy} variant="icon" size="sm" aria-label={label}
                ><Icon inline icon={ILucideCopy} aria-hidden /></Button
            ></slot
        ></Tooltip
    ></span
>

<style lang="scss">
    .copy-button {
        display: contents;
        &:hover {
            --icon-color: var(--body-color);
        }
    }
</style>
