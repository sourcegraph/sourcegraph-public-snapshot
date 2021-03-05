import { storiesOf } from '@storybook/react'

type StoryCategory = 'web' | 'shared' | 'browser' | 'branded'

// TODO: consider using the latest API for adding stories
// https://storybook.js.org/docs/react/writing-stories/introduction
export const addStory = (category: StoryCategory, kind: string): ReturnType<typeof storiesOf> =>
    storiesOf(`${category}/${kind}`, module)
