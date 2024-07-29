<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import Button from '$lib/wildcard/Button.svelte'

    export let open: boolean
    export let handleDismiss: () => void

    let root: HTMLDialogElement
    $: if (open) {
        root?.showModal()
    } else {
        root?.close()
    }
</script>

<dialog bind:this={root}>
    <div class="content">
        <div class="logo"><Icon icon={ISgMark} /></div>
        <h1><span>You've activated a better, faster experience</span> ⚡</h1>
        <p class="subtitle">
            Get ready for a new Code Search experience: rewritten from the ground-up for performance to empower your
            workflow.
        </p>
        <div class="features">
            <div>
                <Icon icon={ILucideFileDiff} />
                <h5>New in-line diff view</h5>
                <p>Easily compare commits and see how a file changed over time, all in-line</p>
            </div>
            <div>
                <Icon icon={ILucideNetwork} />
                <h5>Revamped code navigation</h5>
                <p>Quickly find a list of references of a given symbol, or immediately jump to the definition</p>
            </div>
            <div>
                <Icon icon={ILucideScanSearch} />
                <h5>Reworked fuzzy finder</h5>
                <p>Find files and symbols quickly and easily with our whole new fuzzy finder.</p>
            </div>
        </div>
        <div class="cta">
            <Button on:click={handleDismiss}>Awesome. I’m ready to use it!</Button>
            <a href="TODO">Read release notes</a>
        </div>
        <p class="footer"> You can opt out at any time by using the toggle at the top of the screen.</p>
    </div>
</dialog>

<style lang="scss">
    dialog {
        width: 80vw;
        height: 80vh;
        border-radius: 0.75rem;
        border: 1px solid var(--border-color);
        padding: 0;
        overflow: hidden;
        background-color: var(--color-bg-1);

        box-shadow: var(--fuzzy-finder-shadow);

        &::backdrop {
            background: var(--fuzzy-finder-backdrop);
        }

        @media (--mobile) {
            border-radius: 0;
            border: none;
            position: fixed;
            width: 100vw;
            height: 100vh;
            max-height: 100vh;
            max-width: 100vw;
        }
    }

    .content {
        // TODO: import this from shadcn color library (once it exists)
        :global(.theme-light) & {
            --color-text-subtle: var(--text-body);
        }
        :global(.theme-dark) & {
            --color-text-subtle: #a6b6d9;
        }

        display: flex;
        gap: 1rem;
        flex-direction: column;

        .logo {
            grid-row: 1;
            grid-column: 1 / -1;
            --icon-color: initial;
            --icon-size: 32px;
        }

        h1 {
            margin: 0;
            grid-row: 2;
            grid-column: 1 / -1;
            text-wrap: balance;

            span {
                background: linear-gradient(90deg, #00cbec 0%, #a112ff 48.53%, #ff5543 97.06%);
                color: transparent;
                background-clip: text;
            }
        }

        .subtitle {
            grid-row: 3;
            grid-column: 1 / -1;
            font-size: var(--font-size-large);
            font-weight: 500;
            color: var(--color-text-subtle);
        }

        .features {
            display: grid;
            max-width: 800px;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 0.5rem 0.75rem;

            > div {
                display: grid;
                grid-template-columns: min-content auto;
                gap: 0.25rem 0.75rem;
                :global([data-icon]) {
                    --icon-size: 20px;
                    grid-column: 1;
                    grid-row: 1;
                }
                h5 {
                    all: unset;
                    font-weight: 600;

                    grid-column: 2;
                    grid-row: 1;
                }
                p {
                    all: unset;
                    font-size: var(--font-size-small);
                    font-weight: 400;
                    color: var(--color-text-subtle);

                    grid-column: 2;
                    grid-row: 2;
                }
            }
        }

        .cta {
            grid-column: 1 / -1;
            display: flex;
            gap: 1rem;
            align-items: center;
        }

        .footer {
            grid-column: 1 / -1;
            color: var(--text-muted);
            font-size: var(--font-size-small);
            font-weight: 400;
        }
    }
</style>
