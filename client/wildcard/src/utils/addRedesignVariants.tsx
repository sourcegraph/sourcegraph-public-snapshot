import React, { ReactElement } from 'react'
import { REDESIGN_CLASS_NAME } from '../hooks/useRedesignToggle'

/**
 *
 * To be able to rely on screenshot tests for redesigned components this wrapper function
 * renders copy of the story wrapped into a div with redesign className next to the initial story.
 *
 * See usage example in 'client/wildcard/src/components/PageSelector/PageSelector.story.tsx'
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
