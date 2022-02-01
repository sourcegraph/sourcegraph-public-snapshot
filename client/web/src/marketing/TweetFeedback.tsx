import TwitterIcon from 'mdi-react/TwitterIcon'
import * as React from 'react'

import { Button } from '@sourcegraph/wildcard'

export interface TweetFeedbackProps {
    score: number
    feedback: string
}

const SCORE_TO_TWEET = 9

export const TweetFeedback: React.FunctionComponent<TweetFeedbackProps> = ({ feedback, score }) => {
    if (score >= SCORE_TO_TWEET) {
        const url = new URL('https://twitter.com/intent/tweet')
        url.searchParams.set('text', `After using @sourcegraph: ${feedback}`)
        return (
            <>
                <p className="mt-2">
                    One more favor, could you share your feedback on Twitter? We'd really appreciate it!
                </p>
                <Button
                    className="d-inline-block mt-2"
                    href={url.href}
                    target="_blank"
                    rel="noreferrer noopener"
                    variant="primary"
                    as="a"
                >
                    <TwitterIcon className="icon-inline mr-2" />
                    Tweet feedback
                </Button>
            </>
        )
    }

    return null
}
