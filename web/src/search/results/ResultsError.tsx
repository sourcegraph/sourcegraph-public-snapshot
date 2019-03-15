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
        let { error: { message } } = this.props

        const match = message.match(/Did you mean `.*`?/)
        if (!match) {
            return upperFirst(message)
        }

        const error = message.slice(0, match.index)
        const suggestion = message.slice(match.index)
        const query = buildSearchURLQuery(suggestion)

        return (
            <>
                {error}
                <Link to={`/search?${query}`}>{suggestion}</Link>
            </>
        )
    }
}
