<script lang="ts">
    import { openFuzzyFinder } from '$lib/fuzzyfinder/FuzzyFinderContainer.svelte'
    import { reposHotkey } from '$lib/fuzzyfinder/keys'
    import type { ExternalServiceKind } from '$lib/graphql-types'
    import Icon from '$lib/Icon.svelte'
    import KeyboardShortcut from '$lib/KeyboardShortcut.svelte'
    import DisplayRepoName from '$lib/repo/DisplayRepoName.svelte'
    import { getHumanNameForCodeHost } from '$lib/repo/shared/codehost'
    import { getButtonClassName } from '$lib/wildcard/Button'
    import DropdownMenu from '$lib/wildcard/menu/DropdownMenu.svelte'
    import MenuButton from '$lib/wildcard/menu/MenuButton.svelte'
    import MenuLink from '$lib/wildcard/menu/MenuLink.svelte'
    import MenuSeparator from '$lib/wildcard/menu/MenuSeparator.svelte'

    export let repoName: string
    export let repoURL: string

    export let externalURL: string | undefined
    export let externalServiceKind: ExternalServiceKind | undefined
</script>

<DropdownMenu triggerButtonClass="{getButtonClassName({ variant: 'text' })} triggerButton">
    <h2 slot="trigger">
        <DisplayRepoName {repoName} kind={externalServiceKind} />
    </h2>

    <MenuLink href={repoURL}>
        <div class="menu-item">
            <Icon icon={ILucideHome} inline />
            <span>Go to repository root</span>
            <KeyboardShortcut shortcut={{ key: 'ctrl+backspace', mac: 'cmd+backspace' }} />
        </div>
    </MenuLink>
    <MenuButton on:click={() => openFuzzyFinder('repos')}>
        <div class="menu-item">
            <Icon icon={ILucideRepeat} inline />
            <span>Switch repo</span>
            <KeyboardShortcut shortcut={reposHotkey} />
        </div>
    </MenuButton>
    <MenuLink href="{repoURL}/-/settings">
        <div class="menu-item">
            <Icon icon={ILucideSettings} inline />
            <span>Settings</span>
        </div>
    </MenuLink>
    {#if externalURL}
        <MenuSeparator />
        <MenuLink href={externalURL} target="_blank" rel="noreferrer noopener">
            <div class="code-host-item">
                <small>
                    {#if externalServiceKind}
                        Hosted on {getHumanNameForCodeHost(externalServiceKind)}
                    {:else}
                        View on code host
                    {/if}
                </small>
                <div class="repo-name">
                    <DisplayRepoName {repoName} kind={externalServiceKind} />
                </div>
                <div class="external-link-icon">
                    <Icon icon={ILucideExternalLink} aria-hidden />
                </div>
            </div>
        </MenuLink>
    {/if}
</DropdownMenu>

<style lang="scss">
    :global(.triggerButton) {
        border-radius: 0;
    }
    h2 {
        margin: 0;
        font-size: var(--font-size-base);
        :global([data-path-container]) {
            font-weight: 500;
        }
        :global([data-slash]) {
            font-weight: 400;
        }
    }

    .menu-item {
        --icon-color: currentColor;

        display: flex;
        gap: 0.5rem;
        min-width: 20rem;
        align-items: center;
        color: var(--color-text);

        :global(kbd) {
            margin-left: auto;
        }
    }

    .code-host-item {
        --icon-color: currentColor;

        display: grid;
        gap: 0.25rem;
        align-items: center;
        grid-template-columns: 1fr min-content;
        grid-template-rows: min-content min-content;

        small {
            color: var(--text-muted);
            grid-column: 1;
            grid-row: 1;
        }

        div.repo-name {
            grid-column: 1;
            grid-row: 2;
            display: flex;
            gap: 0.5rem;
            align-items: center;
        }

        div.external-link-icon {
            grid-column: 2;
            grid-row: 1 / span 2;
            --icon-size: 1rem;
        }
    }
</style>
