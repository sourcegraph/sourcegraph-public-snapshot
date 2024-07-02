import React from 'react'

import { Navigate, useLocation } from 'react-router-dom'

import { getReturnTo } from './SignInSignUpCommon'

export const PostSignUpPage: React.FunctionComponent = () => {
    const location = useLocation()
    const returnTo = getReturnTo(location)

    // Redirects Cody PLG users without asking
    const params = new URLSearchParams()
    params.set('returnTo', returnTo)

    const navigateTo = '/sign-in?' + params.toString()

    return <Navigate to={navigateTo.toString()} replace={true} />
}
