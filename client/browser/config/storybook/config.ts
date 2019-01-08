import { configure } from '@storybook/react'

// Automatically import all files ending in *.stories.tsx?.
const req = require.context('../../stories', true, /\.tsx?$/)
function loadStories(): void {
    for (const filename of req.keys()) {
        req(filename)
    }
}

configure(loadStories, module)
