<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'

    import LoadingSpinner from './LoadingSpinner.svelte'

    export const meta = {
        component: LoadingSpinner,
    }
</script>

<Story name="Default">
    <h3>Default</h3>
    <div class="wrapper">
        <LoadingSpinner />
    </div>

    <h3>--size="2rem"</h3>
    <div class="wrapper">
        <LoadingSpinner --size="2rem" />
    </div>

    <h3>center={false}</h3>
     <div class="wrapper">
        <LoadingSpinner center={false} />
    </div>

    <h3>inline={true}</h3>
    <div class="wrapper">
        <LoadingSpinner inline />
    </div>
    <button>
        <LoadingSpinner inline />
        <span>Loading...</span>
    </button>
    <br />
    <button style="font-size: 48px">
        <LoadingSpinner inline />
        <span>Loading...</span>
    </button>
</Story>

<style lang="scss">
    .wrapper {
        display: flex;
        margin: 1rem;
        width: 10rem;
        height: 10rem;
        background-color: lightblue;
    }

    button span {
        vertical-align: middle;
    }
</style>

