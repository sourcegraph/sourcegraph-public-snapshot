<!--
    @component
    A component that displays a list of file or directory actions. It is used in
    the file header of the file tree and file page components.

    Entries should mark their label with the `data-action-label` attribute to
    ensure that it is hidden on small screens.

    Menu entries should use the {@link MenuLink} and {@link MenuButton} components.

    @slot - The actions that are always displayed
    @slot menu - The actions that are displayed in a dropdown menu (if any)
-->
<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import { DropdownMenu } from '$lib/wildcard'
    import { getButtonClassName } from '$lib/wildcard/Button'
    import { mdiDotsHorizontal } from '@mdi/js'
</script>

<div class="root">
    <slot />
    {#if $$slots.menu}
        <div class="divider" />
        <div class="more">
            <DropdownMenu triggerButtonClass={getButtonClassName({ variant: 'icon' })} aria-label="Show more actions">
                <svelte:fragment slot="trigger">
                    <Icon svgPath={mdiDotsHorizontal} inline />
                </svelte:fragment>
                <slot name="menu" />
            </DropdownMenu>
        </div>
    {/if}
</div>

<style lang="scss">
    .root {
        display: flex;
        gap: 1rem;
        align-items: baseline;
        margin-right: 0.5rem;
        --color: var(--icon-color);

        .divider {
            border-left: 1px solid var(--border-color);
            align-self: stretch;
        }

        .more {
            align-self: center;
        }

        @container fileheader (width < 50cqw) {
            :global([data-action-label]) {
                display: none;
            }
        }
    }
</style>
