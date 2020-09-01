import React from 'react'
import { BannerPuzzleIcon } from './icons'

export const ExtensionBanner: React.FunctionComponent = React.memo(() => (
    <>
        <hr className="extension-banner__divider" />

        <div className="extension-banner">
            <div className="extension-banner__card">
                <BannerPuzzleIcon />
                <div className="extension-banner__text-container">
                    <h3>Create your own extension</h3>
                    <p>
                        You can improve your workflow by creating custom extensions. Read the Sourcegraph Docs for
                        details about writing and publishing.
                    </p>
                    <a className="btn btn-primary mt-2" href="https://docs.sourcegraph.com/extensions/authoring">
                        Learn more
                    </a>
                </div>
            </div>
        </div>
    </>
))
