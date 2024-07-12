<script lang="ts">
    import { openFuzzyFinder } from '$lib/fuzzyfinder/FuzzyFinderContainer.svelte'
    import { reposHotkey } from '$lib/fuzzyfinder/keys'
    import Icon from '$lib/Icon.svelte'
    import KeyboardShortcut from '$lib/KeyboardShortcut.svelte'
    import type { DisplayRepoName_ExternalLink } from '$lib/repo/shared/DisplayRepoName.gql'
    import DisplayRepoName from '$lib/repo/shared/DisplayRepoName.svelte'
    import { getHumanNameForExternalService } from '$lib/repo/shared/externalService'
    import { getButtonClassName } from '$lib/wildcard/Button'
    import DropdownMenu from '$lib/wildcard/menu/DropdownMenu.svelte'
    import MenuButton from '$lib/wildcard/menu/MenuButton.svelte'
    import MenuLink from '$lib/wildcard/menu/MenuLink.svelte'
    import MenuSeparator from '$lib/wildcard/menu/MenuSeparator.svelte'

    export let repoName: string
    export let repoURL: string
    export let externalLinks: DisplayRepoName_ExternalLink[]
</script>

<DropdownMenu triggerButtonClass="{getButtonClassName({ variant: 'text' })} triggerButton">
    <div slot="trigger" class="trigger">
        <h2>
            <DisplayRepoName {repoName} {externalLinks} />
        </h2>
    </div>

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
    {#if externalLinks.length > 0}
        <MenuSeparator />
        {#each externalLinks as externalLink (externalLink.url)}
            <MenuLink href={externalLink.url} target="_blank" rel="noreferrer noopener">
                <div class="code-host-item">
                    <small>
                        {#if externalLink.serviceKind}
                            Hosted on {getHumanNameForExternalService(externalLink.serviceKind)}
                        {:else}
                            View on code host
                        {/if}
                    </small>
                    <div class="repo-name">
                        <DisplayRepoName {repoName} externalLinks={[externalLink]} />
                    </div>
                    <div class="external-link-icon">
                        <Icon icon={ILucideExternalLink} aria-hidden />
                    </div>
                </div>
            </MenuLink>
        {/each}
    {/if}
</DropdownMenu>

<style lang="scss">
    :global(.triggerButton) {
        border-radius: 0;
    }
    .trigger {
        --icon-color: currentColor;

        display: flex;
        align-items: center;
        gap: 0.5rem;
        white-space: nowrap;

        h2 {
            font-size: var(--font-size-large);
            font-weight: 500;
            margin: 0;
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
