import React from 'react'

import classNames from 'classnames'

import { Link } from '@sourcegraph/wildcard'

import styles from './AnnouncementBlurb.module.scss'

export interface AnnouncementBlurb {
    text: string;
    link?: LinkProps;
    displayUntil: Date;
}

interface LinkProps {
    text: string;
    link: string;
}

export const AnnouncementBlurb: React.FunctionComponent<AnnouncementBlurb> = ({ text, link, displayUntil }) => (
    // Display only if today is not past display cut off date
    <div className={classNames(new Date() < displayUntil ? 'd-block' : 'd-none',styles.saturnGradientBorder)}>
        <div className="py-2 px-3">
            {text}
            {link && (
                <Link to={link.link} className={classNames('ml-1', styles.link)} target="_blank" rel="noopener noreferrer">
                    {link.text}
                </Link>
            )}
        </div>
    </div>
)
