<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import Popover from '$lib/Popover.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { mdiClose, mdiLightningBolt } from '@mdi/js'
    import { SearchMode, type QueryStateStore } from '../state'

    export let queryState: QueryStateStore

    $: smartEnabled = $queryState.searchMode === SearchMode.SmartSearch
</script>

<Popover let:registerTrigger let:toggle>
    <Tooltip tooltip="Smart search {smartEnabled ? 'enabled' : 'disabled'}">
        <button
            class="toggle icon"
            type="button"
            class:active={smartEnabled}
            on:click={() => toggle()}
            use:registerTrigger
        >
            <Icon svgPath={mdiLightningBolt} inline />
        </button>
    </Tooltip>
    <div slot="content" class="popover-content" let:toggle>
        {@const delayedClose = () => setTimeout(() => toggle(false), 100)}
        <div class="d-flex align-items-center px-3 py-2">
            <h4 class="m-0 mr-auto">SmartSearch</h4>
            <button class="icon" type="button" on:click={() => toggle(false)}>
                <Icon svgPath={mdiClose} inline />
            </button>
        </div>
        <div>
            <label class="d-flex align-items-start">
                <input
                    type="radio"
                    name="mode"
                    value="smart"
                    checked={smartEnabled}
                    on:click={() => {
                        queryState.setMode(SearchMode.SmartSearch)
                        delayedClose()
                    }}
                />
                <span class="d-flex flex-column ml-1">
                    <span>Enable</span>
                    <small class="text-muted"
                        >Suggest variations of your query to find more results that may relate.</small
                    >
                </span>
            </label>
            <label class="d-flex align-items-start">
                <input
                    type="radio"
                    name="mode"
                    value="precise"
                    checked={!smartEnabled}
                    on:click={() => {
                        queryState.setMode(SearchMode.Precise)
                        delayedClose()
                    }}
                />
                <span class="d-flex flex-column ml-1">
                    <span>Disable</span>
                    <small class="text-muted">Only show results that previsely match your query.</small>
                </span>
            </label>
        </div>
    </div>
</Popover>

<style lang="scss">
    button.toggle {
        width: 1.5rem;
        height: 1.5rem;
        cursor: pointer;
        border-radius: var(--border-radius);
        display: flex;
        align-items: center;
        justify-content: center;

        &.active {
            background-color: var(--primary);
            color: var(--light-text);
        }

        :global(svg) {
            transform: scale(1.172);
        }
    }

    .divider {
        width: 1px;
        height: 1rem;
        background-color: var(--border-color-2);
        margin: 0 0.5rem;
    }

    button.icon {
        padding: 0;
        margin: 0;
        border: 0;
        background-color: transparent;
        cursor: pointer;
    }

    .popover-content {
        input {
            margin-left: 0;
        }

        label {
            max-width: 17rem;
            display: flex;
            cursor: pointer;
            padding: 0.5rem 1rem;
            border-top: 1px solid var(--border-color);
        }
    }
</style>
