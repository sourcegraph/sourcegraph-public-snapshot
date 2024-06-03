<script lang="ts">
    import type { HTMLInputAttributes } from 'svelte/elements'

    type $$Props = HTMLInputAttributes
</script>

<input {...$$restProps} type="checkbox" role="switch" on:change />

<style lang="scss">
    input {
        --thumb-size: 1rem;
        --thumb-color: var(--body-bg);
        --thumb-position: 0%;
        --thumb-transition-duration: 0.2s;

        --track-size: calc(var(--thumb-size) * 2);
        --track-padding: 2px;
        --track-color-inactive: var(--gray-05);
        --track-color-active: var(--primary);

        appearance: none;
        border: none;
        box-sizing: content-box !important;
        cursor: pointer;

        inline-size: var(--track-size);
        block-size: var(--thumb-size);
        padding: var(--track-padding) !important;
        background: var(--track-color-inactive);
        border-radius: var(--track-size);

        flex-shrink: 0;
        display: grid;
        align-items: center;
        grid: [track] 1fr / [track] 1fr;

        :global(.theme-dark) & {
            --track-color-inactive: var(--gray-07);
        }

        &::before {
            content: '';
            grid-area: track;
            inline-size: var(--thumb-size);
            block-size: var(--thumb-size);
            background: var(--thumb-color);
            border-radius: 50%;
            transform: translateX(var(--thumb-position));

            @media (prefers-reduced-motion: no-preference) {
                transition: transform var(--thumb-transition-duration) ease;
            }
        }

        &:checked {
            background: var(--track-color-active);
            --thumb-position: calc(var(--track-size) - 100%);
        }

        &:disabled {
            cursor: not-allowed;
            --thumb-color: transparent;

            &::before {
                cursor: not-allowed;
                box-shadow: inset 0 0 0 2px hsl(0 0% 100% / 50%);

                :global(.theme-dark) & {
                    box-shadow: inset 0 0 0 2px hsl(0 0% 0% / 50%);
                }
            }
        }
    }
</style>
