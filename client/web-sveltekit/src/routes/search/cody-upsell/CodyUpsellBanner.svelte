<script lang="ts">
    import CodyIcon from '$lib/icons/Cody.svelte'

    import MultiLineCompletion from './MultiLineCompletion.svelte'
</script>

<section class="upsell">
    <div class="meta">
        <div class="logo"><CodyIcon /></div>
        <h2 class="title">Introducing Cody: your new AI coding assistant.</h2>
        <p class="description">
            Cody autocompletes single lines, or entire code blocks, in any programming language, keeping all of your
            company’s codebase in mind.
        </p>
        <a href="/cody">Explore Cody</a>
    </div>
    <div class="image"><MultiLineCompletion /></div>
</section>

<style lang="scss">
    .upsell {
        padding: 1.75rem 2.5rem;
        padding-right: 1rem;
        margin-top: 4rem;
        max-width: 60rem;
        border-radius: 0.5rem;
        display: grid;
        grid-template-columns: 1fr 1.5fr;
        gap: 1rem;

        @media (--sm-breakpoint-down) {
            grid-template-columns: 1fr;
        }

        .meta {
            align-self: center;

            .logo {
                width: 2.5rem;
                height: 2.5rem;
                margin-bottom: 1rem;
            }

            .title {
                color: var(--text-title);
                font-size: 1.5rem;
                font-weight: 500;
                margin-bottom: 0.75rem;
            }

            .description {
                color: var(--text-body);
                font-size: 0.9375rem;
                margin-bottom: 1.25rem;
            }
        }

        .image {
            filter: drop-shadow(-7px -16px 32px #a112ff24);
            width: 100%;
        }
    }
</style>
