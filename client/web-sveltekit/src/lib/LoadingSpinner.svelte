<script lang="ts">
    export let inline = false
    export let center = true
</script>

<div class:center>
    <div class="loading-spinner" class:inline aria-label="loading" aria-live="polite" />
</div>

<style lang="scss">
    $size: var(--size, 1rem);

    .center {
        display: flex;
        flex-direction: column;
        align-items: center;
        flex: 1;
        justify-content: center;
    }

    .loading-spinner {
        :global(.theme-light) & {
            --loading-spinner-outer-color: var(--gray-05);
            --loading-spinner-inner-color: var(--gray-08);
        }
        :global(.theme-dark) & {
            --loading-spinner-outer-color: var(--gray-07);
            --loading-spinner-inner-color: var(--white);
        }

        width: $size;
        height: $size;
        &.inline {
            $inlineSize: #{(16 / 14)}em;
            width: $inlineSize;
            height: $inlineSize;

            vertical-align: bottom;
            display: inline-flex;
            align-items: center;
        }

        border-radius: 50%;
        animation: loading-spinner-spin 1s linear infinite;
        border: 2px solid var(--loading-spinner-outer-color, rgba(0, 0, 0, 0.3));
        border-top: 2px solid var(--loading-spinner-inner-color, rgba(0, 0, 0, 1));
    }

    @keyframes loading-spinner-spin {
        0% {
            transform: rotate(0deg);
        }
        100% {
            transform: rotate(360deg);
        }
    }
</style>
