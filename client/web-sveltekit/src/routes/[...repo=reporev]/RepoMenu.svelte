<script lang="ts">
    import { openFuzzyFinder } from '$lib/fuzzyfinder/FuzzyFinderContainer.svelte'
    import { reposHotkey } from '$lib/fuzzyfinder/keys'
    import Icon from '$lib/Icon.svelte'
    import KeyboardShortcut from '$lib/KeyboardShortcut.svelte'
    import { getHumanNameForCodeHost } from '$lib/repo/shared/codehost'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import { getButtonClassName } from '$lib/wildcard/Button'
    import DropdownMenu from '$lib/wildcard/menu/DropdownMenu.svelte'
    import MenuButton from '$lib/wildcard/menu/MenuButton.svelte'
    import MenuLink from '$lib/wildcard/menu/MenuLink.svelte'
    import MenuSeparator from '$lib/wildcard/menu/MenuSeparator.svelte'

    export let repoName: string
    export let displayRepoName: string
    export let repoURL: string

    export let externalURL: string | undefined
    export let externalServiceKind: string | undefined
</script>

<DropdownMenu triggerButtonClass="{getButtonClassName({ variant: 'text' })} triggerButton">
    <svelte:fragment slot="trigger">
        <div class="trigger">
            <CodeHostIcon repository={repoName} codeHost={externalServiceKind} />
            <h2>
                {#each displayRepoName.split('/') as segment, i}
                    {#if i > 0}<span class="slash">/</span>{/if}{segment}
                {/each}
            </h2>
        </div>
    </svelte:fragment>

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
                    <CodeHostIcon repository={repoName} codeHost={externalServiceKind} />
                    <span>{displayRepoName}</span>
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

            .slash {
                font-weight: 400;
                color: var(--text-muted);
                margin: 0.25em;
                letter-spacing: -0.25px;
            }
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
