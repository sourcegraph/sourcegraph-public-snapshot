<script lang="ts">
    import { mdiMerge } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'

    export let name: string = 'sourcegraph'
    export let codeHostIcon: string = mdiMerge
</script>

<div class="container">
    <div class="code-host-icon">
        <Icon svgPath={codeHostIcon} />
    </div>
    <div class="right-side">
        <h4 class="repo-name">sourcegraph<span>&nbsp;/&nbsp;</span>sourcegraph</h4>
        <div class="switch"><small>Switch</small></div>
    </div>
</div>

<style lang="scss">
    span {
        color: var(--text-muted);
    }
    h4 {
        font-weight: normal;
    }

    .container {
        display: flex;
        flex-flow: row nowrap;
        align-items: center;
        justify-content: space-between;
        padding: 0.25rem 0.5rem;
    }

    .code-host-icon {
        width: 1.5rem;
    }

    .right-side {
        display: flex;
        flex-flow: column nowrap;
        justify-items: flex-start;
        margin-left: 0.5rem;

        .repo-name {
            text-align: left;
            margin: 0rem;
        }

        .switch {
            color: var(--text-muted);
            text-align: left;
        }
    }
</style>
