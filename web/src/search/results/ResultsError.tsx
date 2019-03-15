import { upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { ErrorLike } from '../../../../shared/src/util/errors'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'

interface Props {
    error: ErrorLike
}

export class ResultsError extends React.Component<Props> {
    public render(): React.ReactNode {
        return (
            <div className="alert alert-warning">
                <AlertCircleIcon className="icon-inline" />
                {this.renderMessage()}
            </div>
        )
    }

    private renderMessage(): React.ReactNode {
        const {
            error: { message },
        } = this.props

        // TODO: Send more robust error response from backend to prevent the need to string match.
        const match = message.match(/Did you mean `(.*?)`/)
        if (!match) {
            return upperFirst(message)
        }

        const suggestion = match[1]

        const [before] = message.split(suggestion)

        const query = buildSearchURLQuery(suggestion)

        const [firstLine] = before.split('Did you mean')

        return (
            <>
                {firstLine}
                <br />
                <br />
                {'Did you mean `'}
                <Link to={`/search?${query}`}>{suggestion}</Link>
                {'`?'}
            </>
        )
    }
}
