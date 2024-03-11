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

    export let variant: 'info' | 'warning' | 'danger' | 'success'
    export let partialStorageKey: string | null

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
    <Alert class={classNames('root', className)} variant={variant} sizeVariant='slim'>

        <div class='content'>
            <slot/>
        </div>

        {#if partialStorageKey}
            <div class="button-wrapper">
                <Button
                    variant="icon"
                    aria-label="Dismiss alert"
                    class="close-button"
                    on:click={handleDismissClick}
                >
                    <Icon aria-hidden={true} svgPath={mdiClose}/>
                </Button>
            </div>
        {/if}
    </Alert>
{/if}

<style lang="scss">
  .root {
    display: flex;
    align-items: center;
    padding-right: 0.5rem;

      &:last-child {
          border-bottom-width: 0;
      }
  }

  .content {
    display: flex;
    flex: 1 1 auto;
    line-height: (20/14);
  }

  .button-wrapper {
    align-self: flex-start;
    color: var(--icon-color);
  }
</style>
