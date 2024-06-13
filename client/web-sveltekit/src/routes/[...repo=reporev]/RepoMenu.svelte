<script lang="ts">
    import { openFuzzyFinder } from '$lib/fuzzyfinder/FuzzyFinderContainer.svelte'
    import { reposHotkey } from '$lib/fuzzyfinder/keys'
    import Icon from '$lib/Icon.svelte'
    import KeyboardShortcut from '$lib/KeyboardShortcut.svelte'
    import { getHumanNameForCodeHost } from '$lib/repo/shared/codehost'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
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

<DropdownMenu triggerButtonClass="trigger">
    <svelte:fragment slot="trigger">
        <CodeHostIcon repository={repoName} codeHost={externalServiceKind} />
        <h2>
            {#each displayRepoName.split('/') as segment, i}
                {#if i > 0}<span class="slash">/</span>{/if}{segment}
            {/each}
        </h2>
    </svelte:fragment>

    <MenuLink href={repoURL}>
        <div class="menu-item">
            <Icon icon={ILucideHome} inline />
            <span>Go to repository root</span>
            <KeyboardShortcut shortcut={{ key: 'ctrl+backspace', mac: 'cmd+backspace' }} />
        </div>
    </MenuLink>
    <MenuButton class="menu-item" on:click={() => openFuzzyFinder('repos')}>
        <div class="menu-item">
            <Icon icon={ILucideRepeat} inline />
            <span>Switch repo</span>
            <KeyboardShortcut shortcut={reposHotkey} />
        </div>
    </MenuButton>
    <MenuLink href={repoURL + '/-/settings'} class="menu-item">
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
                <div class="">
                    <CodeHostIcon repository={repoName} codeHost={externalServiceKind} />
                    <span>{displayRepoName}</span>
                </div>
            </div>
        </MenuLink>
    {/if}
</DropdownMenu>

<style lang="scss">
    :global(.trigger) {
        all: unset;

        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0 1rem;
        white-space: nowrap;

        &:hover {
            background-color: var(--secondary-2);
        }

        :global(h2) {
            font-size: var(--font-size-large);
            font-weight: 500;
            margin: 0;

            .slash {
                color: var(--text-muted);
                margin: 0.25rem;
            }
        }
    }

    :global(.menu-item) {
        display: flex;
        gap: 0.5rem;
        min-width: 20rem;
        align-items: center;
        color: var(--color-text);
        --icon-color: currentColor;

        :global(kbd) {
            margin-left: auto;
        }
    }

    .code-host-item {
        display: grid;
        flex-direction: column;
        gap: 0.25rem;

        small {
            color: var(--text-muted);
        }
    }
</style>
