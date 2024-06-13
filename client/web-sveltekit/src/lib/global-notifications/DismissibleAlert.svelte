<script lang="ts" context="module">
    function dismissAlert(key: string): void {
        localStorage.setItem(storageKeyForPartial(key), 'true')
    }

    function isAlertDismissed(key: string): boolean {
        return localStorage.getItem(storageKeyForPartial(key)) === 'true'
    }

    function storageKeyForPartial(partialStorageKey: string): string {
        return `DismissibleAlert/${partialStorageKey}/dismissed`
    }

    export function clearDismissedAlertsState_TEST_ONLY(...partialStorageKeys: string[]): void {
        for (const partialStorageKey of partialStorageKeys) {
            localStorage.removeItem(storageKeyForPartial(partialStorageKey))
        }
    }
</script>

<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import { Alert, Button } from '$lib/wildcard'

    export let variant: 'info' | 'warning' | 'danger' | 'success'
    export let partialStorageKey: string | null

    // Local state
    let dismissed = partialStorageKey ? isAlertDismissed(partialStorageKey) : false

    // Callback handlers
    function handleDismissClick() {
        if (partialStorageKey) {
            dismissAlert(partialStorageKey)
        }

        dismissed = true
    }
</script>

{#if !dismissed}
    <Alert {variant} size="slim">
        <div class="content-wrapper">
            <div class="content">
                <slot />
            </div>

            {#if partialStorageKey}
                <div class="button-wrapper">
                    <Button variant="icon" aria-label="Dismiss alert" on:click={handleDismissClick}>
                        <Icon aria-hidden inline icon={ILucideX} />
                    </Button>
                </div>
            {/if}
        </div>
    </Alert>
{/if}

<style lang="scss">
    .content {
        display: flex;
        flex: 1 1 auto;
        line-height: (20/14);
    }

    .content-wrapper {
        display: flex;
        align-items: center;
        overflow: hidden;
    }

    .button-wrapper {
        align-self: flex-start;
        color: var(--icon-color);
    }
</style>
