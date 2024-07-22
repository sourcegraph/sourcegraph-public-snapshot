<script lang="ts">
    import type { HTMLButtonAttributes } from 'svelte/elements'

    type $$Props = {
        on: boolean
    } & HTMLButtonAttributes

    export let on: boolean
</script>

<button type="button" class="toggle" role="switch" aria-checked={on} {...$$restProps} on:click>
    <span class="bar" class:bar--on={on} />
    <span class="knob" class:knob--on={on} />
</button>

<style lang="scss">
    .toggle {
        --toggle-width: 2rem;
        --toggle-bar-bg: var(--icon-color);
        --toggle-bar-bg-on: var(--primary);
        --toggle-knob-bg: var(--body-bg);
        --toggle-knob-bg-on: var(--body-bg);
        --toggle-bar-opacity: 1;
        --toggle-bar-focus-opacity: 1;
        --toggle-knob-disabled-opacity: 1;
        --toggle-bar-focus-box-shadow: 0 0 0 1px var(--body-bg), 0 0 0 0.1875rem var(--primary-2);

        background: none;
        border: none;
        outline: none !important;
        padding: 0;
        position: relative;
        width: var(--toggle-width);

        height: 1rem;
        display: inline-flex;
        align-items: center;

        &:focus-visible {
            /* Move focus style to the rounded bar */
            box-shadow: none;
        }

        &:disabled {
            --toggle-knob-bg: var(--icon-color);
            --toggle-knob-bg-on: var(--icon-color);
            --toggle-bar-bg: var(--input-disabled-bg);
            --toggle-bar-bg-on: var(--input-disabled-bg);
        }

        &:hover:enabled .bar {
            opacity: var(--toggle-bar-focus-opacity);
        }

        &:disabled .knob {
            opacity: var(--toggle-knob-disabled-opacity);
        }

        &:focus-visible .bar {
            box-shadow: var(--toggle-bar-focus-box-shadow);
        }
    }

    .inline-center {
        margin-top: 0.125rem;
    }

    .bar {
        border-radius: 1rem;
        left: 0;
        height: 1rem;
        width: 100%;
        position: absolute;

        opacity: var(--toggle-bar-opacity);
        background-color: var(--toggle-bar-bg);

        transition: all 0.3s;
        transition-property: opacity;

        &--on {
            background-color: var(--toggle-bar-bg-on);
        }
    }

    .knob {
        background-color: var(--toggle-knob-bg);

        border-radius: 0.375rem;
        display: block;

        height: 0.75rem;
        width: 0.75rem;
        left: 0.125rem;

        position: relative;

        &--on {
            background-color: var(--toggle-knob-bg-on);
            transform: translate3d(1rem, 0, 0);
        }
    }
</style>
