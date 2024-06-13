<script lang="ts">
    import type { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'

    import type { Evaluated, ActionContribution } from '$lib/client-api'
    import { isExternalLink } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Button } from '$lib/wildcard'

    type Action = Evaluated<ActionContribution>

    export let actionItem: ActionItemAction

    function processAction(
        action: Action,
        disabled: boolean
    ): { content: string; tooltip: string; icon?: { description: string; url: string } } {
        let content = ''
        let tooltip = ''
        let icon = undefined

        if (action.title) {
            content = action.title
            tooltip = action.description ?? ''
        } else if (action.actionItem) {
            if (action.actionItem.iconURL) {
                icon = { description: action.actionItem.iconDescription ?? '', url: action.actionItem.iconURL }
            }
            if (action.actionItem.label) {
                content = ' ' + action.actionItem.label
            }
            tooltip = action.actionItem.description ?? ''
        } else if (disabled) {
            content = action.disabledTitle ?? ''
        } else {
            if (action.iconURL) {
                icon = { description: action.description ?? '', url: action.iconURL }
            }
            content = (action.category ? `${action.category}: ` : '') + action.title
        }

        return { content, tooltip, icon }
    }

    function actionURL(action: Action): string | undefined {
        if (action.command === 'open') {
            const url = action.commandArguments?.[0]
            return typeof url === 'string' ? url : undefined
        }
        return undefined
    }

    function run(event: MouseEvent) {
        switch (action.command) {
            case 'invokeFunction-new': {
                const args = action.commandArguments || []
                for (const arg of args) {
                    if (typeof arg === 'function') {
                        arg()
                    }
                }
                break
            }
            case 'open': {
                if (action.commandArguments && action.commandArguments.length > 1) {
                    const onSelect = action.commandArguments[1]
                    if (typeof onSelect === 'function') {
                        onSelect(event)
                    }
                }
            }
        }
    }

    $: disabled = actionItem.disabledWhen ?? false
    $: ({ action, altAction } = actionItem)
    $: ({ content, tooltip, icon } = processAction(action, disabled))
    $: url = actionURL(action) ?? (altAction ? actionURL(altAction) : null)
    $: newTabProps = url && isExternalLink(url) ? { target: '_blank', rel: 'noopener noreferrer' } : {}
</script>

<Tooltip {tooltip}>
    {#if !disabled && url}
        <Button variant="secondary" size="sm" {disabled}>
            <svelte:fragment slot="custom" let:buttonClass>
                <a class={buttonClass} href={url} on:click={run} {...newTabProps}>
                    {#if icon}
                        <img src={icon.url} alt={icon.description} />
                    {/if}
                    {content}
                </a>
            </svelte:fragment>
        </Button>
    {:else if action.command}
        <Button variant="secondary" size="sm" {disabled} on:click={run}>
            {#if icon}
                <img src={icon.url} alt={icon.description} />
            {/if}
            {content}
        </Button>
    {:else if action.title === '?'}
        <Icon aria-hidden icon={ILucideCircleHelp} inline />
    {:else}
        {content}
    {/if}
</Tooltip>
