import { configure } from '@storybook/react'

// automatically import all files ending in *.stories.js
const req = require.context('../../stories', true, /\.tsx?$/)
function loadStories(): void {
    for (const filename of req.keys()) {
        req(filename)
    }
}

configure(loadStories, module)
