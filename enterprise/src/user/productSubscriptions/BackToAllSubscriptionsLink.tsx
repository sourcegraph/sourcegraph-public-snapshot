import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import ArrowLeftIcon from 'mdi-react/ArrowLeftIcon'
import React from 'react'
import { Link } from 'react-router-dom'

export const BackToAllSubscriptionsLink: React.SFC<{ user: Pick<GQL.IUser, 'url'> }> = ({ user }) => (
    <Link to={`${user.url}/subscriptions`} className="btn btn-outline-link btn-sm mb-3">
        <ArrowLeftIcon className="icon-inline" /> All subscriptions
    </Link>
)
