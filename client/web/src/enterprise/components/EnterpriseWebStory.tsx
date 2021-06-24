import React from 'react'

import { WebStory, WebStoryProps } from '../../components/WebStory'
import enterpriseWebStyles from '../../enterprise.scss'

/**
 * Wrapper component for enterprise webapp Storybook stories that provides light theme and react-router props.
 * Takes a render function as children that gets called with the props.
 */
export const EnterpriseWebStory: React.FunctionComponent<WebStoryProps> = props => (
    <WebStory {...props} additionalWebStyles={enterpriseWebStyles} />
)
