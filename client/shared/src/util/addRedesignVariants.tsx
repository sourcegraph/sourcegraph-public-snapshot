import React, { ReactElement } from 'react'

import { REDESIGN_CLASS_NAME } from './useRedesignToggle'

/**
 *
 * To rely on screenshot tests for redesigned components in Storybook, this wrapper function
 * renders a copy of the story wrapped into a div with redesign class next to the initial story.
 *
 * @example
 * const { add } = storiesOf('wildcard/PageSelector', module).addDecorator(story => (
 *  <BrandedStory styles={webStyles}>
 *      {() => addRedesignVariants(<div className="container web-content mt-3">{story()}</div>)}
 *  </BrandedStory>
 * ))
 *
 */
export const addRedesignVariants = (story: ReactElement): ReactElement => (
    <>
        {story}
        <hr className="mb-3" />
        <div className={REDESIGN_CLASS_NAME}>
            <div className="badge badge-secondary">Redesign variant</div>
            {story}
        </div>
    </>
)
