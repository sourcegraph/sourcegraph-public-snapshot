import React from 'react'

import { mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { Link, Icon } from '@sourcegraph/wildcard'

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
    <div>
        {/* Display only if today is not past display cut off date */}
        {new Date() < displayUntil && (
            <div className={styles.saturnGradientBorder}>
                <div className="py-2 px-3">
                    {text}
                    {link && (
                        <Link to={link.link} className={classNames('ml-1', styles.link)} target="_blank" rel="noopener noreferrer">
                            {link.text}{' '}
                            <Icon aria-hidden={true} svgPath={mdiOpenInNew} />
                        </Link>
                    )}
                </div>
            </div>
        )}
    </div>
)
