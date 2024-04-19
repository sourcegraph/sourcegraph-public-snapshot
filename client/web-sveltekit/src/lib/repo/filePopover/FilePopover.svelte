<script lang="ts">
    import { mdiLanguageGo } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'

    const CENTER_DOT = '\u00B7' // interpunct

    let org = 'sourcegraph'
    let repo = 'sourcegraph'
    let filePath = 'cmd/frontend/auth'.split('/')
    let fileInfo = `Go ${CENTER_DOT} 58 lines ${CENTER_DOT} 1.43 KB`
</script>

<div class="root">
    <div class="desc">
        <div class="repo-and-path">
            <small>{org}</small>
            <small>/</small>
            <small>{repo}</small>
            <small>{CENTER_DOT}</small>
            {#each filePath as part}
                {#if filePath.indexOf(part) < filePath.length - 1}
                    <small>{part}</small>
                    <small>/</small>
                {:else}
                    <small>{part}</small>
                {/if}
            {/each}
        </div>
        <div class="lang-and-file">
            <div>
                <Icon svgPath={mdiLanguageGo} --color="var(--primary)" />
            </div>
            <div class="file">
                <div class="file-name">auth.go</div>
                <div class="file-info">
                    <small>{fileInfo}</small>
                </div>
            </div>
        </div>
    </div>
</div>

<style lang="scss">
    .root {
        width: fit-content;
        background: var(--body-bg);
        border: 1px solid var(--border-color);
        border-top-right-radius: 8px;
        border-top-left-radius: 8px;

        .desc {
            display: flex;
            flex-flow: column nowrap;
            align-items: center;
            justify-content: center;
            .repo-and-path {
                width: 100%;
                display: flex;
                flex-flow: row nowrap;
                align-items: center;
                justify-content: center;
                gap: 0.25rem;
                padding: 0.5rem 1rem;
                font-family: var(--monospace-font-family);
                font-weight: lighter;
                color: var(--text-muted);
                border-bottom: 1px solid var(--border-color);
            }

            .lang-and-file {
                width: 100%;
                display: flex;
                flex-flow: row nowrap;
                align-items: flex-start;
                justify-content: flex-start;
                padding: 0.5rem 1rem;
                gap: 0.5rem 1rem;
                border-bottom: 1px solid var(--border-color);

                .file {
                    display: flex;
                    flex-flow: column nowrap;
                    align-items: flex-start;
                    justify-content: flex-start;
                    font-family: var(--monospace-font-family);
                    .file-name {
                        color: var(--text-body);
                    }
                    .file-info {
                        color: var(--text-muted);
                    }
                }
            }
        }
    }
</style>
