import React from 'react'

import { Terms } from '@sourcegraph/cody-ui/src/Terms'

import './About.css'

export const About: React.FunctionComponent = () => (
    <div className="inner-container">
        <div className="non-transcript-container">
            <div className="about-container">
                <Terms />
            </div>
        </div>
    </div>
)
