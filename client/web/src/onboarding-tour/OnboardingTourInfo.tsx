import classNames from 'classnames'
import React from 'react'

export const ONBOARDING_TOUR_MARKER = 'onboarding-tour-info-marker'

interface OnboardingTourInfoProps {
    className?: string
}

export const OnboardingTourInfo: React.FunctionComponent<OnboardingTourInfoProps> = ({ className }) => (
    <div className={classNames(ONBOARDING_TOUR_MARKER, className)} />
)
