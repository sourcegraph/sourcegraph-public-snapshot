<script lang="ts">
    import type { ComponentProps } from 'svelte'

    import Icon from '$lib/Icon.svelte'

    export let title: string
    export let icon: ComponentProps<Icon>['icon']
</script>

<div class="root">
    {#if icon || $$slots.icon}
        <div class="icon-wrapper">
            <slot name="icon">
                <Icon {icon} aria-hidden="true" inline --icon-inline-size="4rem" />
            </slot>
        </div>
    {/if}
    <h1>{title}</h1>
    <slot />
</div>

<style lang="scss">
    .root {
        flex: 1 1 auto;
        display: flex;
        flex-direction: column;
        width: 100%;
        max-width: var(--viewport-xl);
        margin: 0 auto;
        min-height: 0;
        align-items: center;
        overflow-y: auto;

        padding-top: 8rem;
    }

    h1 {
        font-size: 2rem;
        margin: 1rem;
        line-height: 1.5;
        font-weight: 400;
        letter-spacing: normal;
    }

    .icon-wrapper {
        border-radius: 100%;
        width: 6rem;
        height: 6rem;
        min-height: 6rem;
        background-color: var(--color-bg-2);
        display: flex;
        align-items: center;
        justify-content: center;

        :global(svg) {
            width: 4rem;
            height: 4rem;
        }
    }
</style>
