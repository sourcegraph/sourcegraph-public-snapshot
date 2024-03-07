<script lang="ts">
    import {renderMarkdown} from '@sourcegraph/common'
    import {settings} from '$lib/stores'
    import Markdown from './Markdown.svelte'
    import DismissibleAlert from './DismissibleAlert.svelte'

    $: settingsMotd = $settings?.motd
    $: notices = $settings?.notices
</script>

<div class="root">
    {#if settingsMotd && Array.isArray(settingsMotd)}
        {#each settingsMotd as motd}
            <DismissibleAlert
                    variant="info"
                    partialStorageKey={`motd.${motd}`}
                    class='alert'
            >
                <Markdown dangerousInnerHTML={renderMarkdown(motd)}/>
            </DismissibleAlert>
        {/each}
    {/if}

    {#if notices && Array.isArray(notices)}
        {#each notices as notice (notice.message)}
            <DismissibleAlert
                    variant="info"
                    partialStorageKey={`notices.${notice.message}`}
                    class='alert'
            >
                <Markdown dangerousInnerHTML={renderMarkdown(notice.message)}/>
            </DismissibleAlert>
        {/each}
    {/if}

    {#if true}
        <DismissibleAlert
                variant="danger"
                partialStorageKey="dev-web-server-alert"
                class='alert'
        >
            <div>
                <strong>Warning!</strong> This build uses data from the proxied API:{' '}
                <a class="proxy-link" target="__blank" href={process.env.SOURCEGRAPH_API_URL}>
                    {process.env.SOURCEGRAPH_API_URL}
                </a>
            </div>
            .
        </DismissibleAlert>
    {/if}
</div>

<style lang="scss">
  .root {
    width: 100%;

    &:empty {
      display: none;
    }
  }

  .root > :global(.alert) {
    width: 100%;
    margin-bottom: 0;
    border-radius: 0;
    border-width: 0 0 1px 0;
    background: var(--alert-icon-background-color);
    padding-left: var(--alert-content-padding);
    border-color: var(--border-color-2);

    &::before,
    &::after {
      display: none;
    }

    &:last-child {
      border-bottom-width: 0;
    }
  }

  .root p {
    margin-bottom: 0;
  }

  /* The trailing after-paragraph/list margin looks unbalanced in MOTDs. */
  p:last-child,
  ul:last-child,
  ol:last-child {
    margin-bottom: 0;
  }

  .proxy-link {
    color: var(--body-color);
  }
</style>
