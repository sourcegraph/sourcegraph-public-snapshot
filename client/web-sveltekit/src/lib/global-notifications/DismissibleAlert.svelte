<script lang="ts" context="module">
    import classNames from 'classnames'

    function dismissAlert(key: string): void {
        localStorage.setItem(storageKeyForPartial(key), 'true')
    }

    function isAlertDismissed(key: string): boolean {
        return localStorage.getItem(storageKeyForPartial(key)) === 'true'
    }

    function storageKeyForPartial(partialStorageKey: string): string {
        return `DismissibleAlert/${partialStorageKey}/dismissed`
    }
</script>

<script lang="ts">
    import {mdiClose} from '@mdi/js'
    import Icon from '$lib/Icon.svelte'
    import {Alert, Button} from '$lib/wildcard'

    export let variant: 'info' | 'danger'
    export let partialStorageKey: string

    let className = '';
    export {className as class};

    // Local state
    let dismissed = partialStorageKey ? isAlertDismissed(partialStorageKey) : false;

    // Callback handlers
    function handleDismissClick() {
        if (partialStorageKey) {
            dismissAlert(partialStorageKey)
        }

        dismissed = true
    }
</script>

{#if !dismissed}
    <Alert class={classNames('root', className)} variant={variant}>

        <div class='content'>
            <slot/>
        </div>

        <Button
                variant="icon"
                aria-label="Dismiss alert"
                on:click={handleDismissClick}
        >
            <Icon aria-hidden={true} svgPath={mdiClose}/>
        </Button>
    </Alert>
{/if}

<style lang="scss">
  .root {
    display: flex;
    align-items: flex-start;
    padding-right: 0.5rem;
  }

  .content {
    display: flex;
    flex: 1 1 auto;
    line-height: (20/14);
  }

  .root > :global(.close-button) {
    height: 1.25rem;
    color: var(--icon-color);
  }
</style>
