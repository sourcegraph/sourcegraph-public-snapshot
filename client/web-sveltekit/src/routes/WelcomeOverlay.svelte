<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import { isLightTheme } from '$lib/theme'
    import Button from '$lib/wildcard/Button.svelte'
    import ProductStatusBadge from '$lib/wildcard/ProductStatusBadge.svelte'

    import WelcomeOverlayScreenshotDark from './WelcomeOverlayScreenshotDark.svelte'
    import WelcomeOverlayScreenshotLight from './WelcomeOverlayScreenshotLight.svelte'

    export let show: boolean
    export let handleDismiss: () => void

    let root: HTMLDialogElement
    $: if (show) {
        root?.showModal()
    } else {
        root?.close()
    }
</script>

<dialog bind:this={root}>
    <div class="content">
        <div class="logo"><Icon icon={ISgMark} /><ProductStatusBadge status="beta" /></div>
        <div class="message">
            <h1><span>You've activated a better, faster experience</span> ⚡</h1>
            <p class="subtitle">
                Get ready for a new Code Search experience: rewritten from the ground-up for performance to empower your
                workflow.
            </p>
        </div>
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
                <!-- TODO: add keyboard shortcut here -->
                <h5>Reworked fuzzy finder</h5>
                <p>Find files and symbols quickly and easily with our whole new fuzzy finder.</p>
            </div>
        </div>
        <div class="cta">
            <div>
                <Button variant="secondary" on:click={() => handleDismiss()}>Awesome. I’m ready to use it!</Button>
                <a href="TODO">Read release notes</a>
            </div>
            <p> You can opt out at any time by using the toggle at the top of the screen.</p>
        </div>
    </div>
    {#if $isLightTheme}
        <WelcomeOverlayScreenshotLight />
    {:else}
        <WelcomeOverlayScreenshotDark />
    {/if}
</dialog>

<style lang="scss">
    dialog {
        width: 80vw;
        border-radius: 0.75rem;
        border: 1px solid var(--border-color);
        padding: 2rem;
        overflow: hidden;
        background-color: var(--color-bg-1);

        box-shadow: var(--fuzzy-finder-shadow);

        &::backdrop {
            backdrop-filter: blur(2px);
        }

        container-type: inline-size;

        @media (--mobile) {
            border-radius: 0;
            border: none;
            position: fixed;
            width: 100vw;
            height: 100vh;
            max-height: 100vh;
            max-width: 100vw;
        }

        > :global(svg) {
            position: absolute;
            right: 0;
            bottom: 0;
            filter: drop-shadow(0px 25px 50px rgba(15, 17, 26, 0.25));
            @container (width < 975px) {
                display: none;
            }
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

        width: calc(100% - 350px);
        @container (width < 975px) {
            width: 100%;
        }

        display: flex;
        gap: 1rem;
        flex-direction: column;

        .logo {
            --icon-color: initial;
            --icon-size: 32px;
            display: flex;
            gap: 1rem;
            align-items: center;
        }

        .message {
            h1 {
                text-wrap: balance;
                span {
                    background: linear-gradient(90deg, #00cbec 0%, #a112ff 48.53%, #ff5543 97.06%);
                    color: transparent;
                    background-clip: text;
                }
            }
        }

        .subtitle {
            margin: 0;
            font-size: var(--font-size-large);
            font-weight: 500;
            color: var(--color-text-subtle);
        }

        .features {
            display: grid;
            max-width: 700px;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 1rem 0.75rem;
            padding: 1rem 0;

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
            display: flex;
            gap: 1rem;
            flex-direction: column;
            div {
                grid-column: 1 / -1;
                display: flex;
                gap: 1rem;
                align-items: center;
            }
            p {
                grid-column: 1 / -1;
                color: var(--text-muted);
                font-size: var(--font-size-small);
                font-weight: 400;
            }
        }
    }
</style>
