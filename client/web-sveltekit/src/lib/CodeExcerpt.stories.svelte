<script lang="ts" context="module">
    import { faker } from '@faker-js/faker'
    import { Story } from '@storybook/addon-svelte-csf'

    import CodeExcerpt from './CodeExcerpt.svelte'

    export const meta = {
        component: CodeExcerpt,
    }

    const code = `
const obj = {
    key:     'value',
    key2:    'value2',
}
`
</script>

<script lang="ts">
    faker.seed(16)
    const plaintextLines = code.trim().split('\n')
    const highlightedHTMLRows = plaintextLines.map(
        (line, index) =>
            `<tr><td class="line" data-line="${
                index + 1
            }" /><td class="code"><span style="color: ${faker.color.rgb()}">${line}</span></td></tr>`
    )
</script>

<Story name="Default">
    <h3>Default</h3>
    <div class="wrapper">
        <CodeExcerpt startLine={1} {plaintextLines} />
    </div>

    <h3>Different start line</h3>
    <div class="wrapper">
        <CodeExcerpt startLine={10} {plaintextLines} />
    </div>

    <h3>Hidden line numbers</h3>
    <div class="wrapper">
        <CodeExcerpt startLine={1} {plaintextLines} hideLineNumbers />
    </div>

    <h3>Collapsed whitespace</h3>
    <div class="wrapper">
        <CodeExcerpt startLine={1} {plaintextLines} collapseWhitespace />
    </div>

    <h3>With highlighted code</h3>
    <div class="wrapper">
        <CodeExcerpt startLine={1} {plaintextLines} {highlightedHTMLRows} />
    </div>
</Story>

<style lang="scss">
    .wrapper {
        background-color: var(--code-bg);
        margin-bottom: 1rem;
    }
</style>
