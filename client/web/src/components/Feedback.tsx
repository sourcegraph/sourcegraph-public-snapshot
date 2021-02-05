import React from 'react'
import { Button } from 'reactstrap'
import { FeedbackIcon } from '../../../shared/src/components/icons'

export const Feedback: React.FunctionComponent = () => (
    <>
        <Button className="d-none d-lg-block" outline={true} color="secondary">
            Feedback
        </Button>
        <Button className="d-lg-none" color="link">
            <FeedbackIcon className="icon-inline" />
        </Button>
    </>
)
