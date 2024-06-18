<script lang="ts">
    import Icon from '$lib/Icon.svelte'

    import MultiLineCompletion from './MultiLineCompletion.svelte'

    export let codyHref = '/cody'
</script>

<section>
    <div class="meta">
        <div class="logo"><Icon icon={ISgCody} aria-label="Cody logo" /></div>
        <h2 class="title">Introducing Cody: your new AI coding assistant.</h2>
        <p class="description">
            Cody autocompletes single lines, or entire code blocks, in any programming language, keeping all of your
            company’s codebase in mind.
        </p>
        <a href={codyHref}>Explore Cody</a>
    </div>
    <div class="image"><MultiLineCompletion /></div>
</section>

<style lang="scss">
    section {
        isolation: isolate;
        padding: 1.75rem 2.5rem;
        display: grid;
        grid-template-columns: 1fr 1.5fr;
        gap: 1rem;

        @media (--sm-breakpoint-down) {
            grid-template-columns: 1fr;
        }

        .meta {
            align-self: center;

            .logo {
                --icon-color: initial;
                --icon-size: 2.5rem;
                margin-bottom: 1rem;
            }

            .title {
                font-size: 1.5rem;
                font-weight: 500;
                margin-bottom: 0.75rem;
            }

            .description {
                margin-bottom: 1.25rem;
            }
        }

        .image {
            filter: drop-shadow(-7px -16px 32px #a112ff24);
            width: 100%;
        }
    }
</style>
