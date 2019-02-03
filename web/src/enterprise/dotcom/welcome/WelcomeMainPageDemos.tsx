import classnames from 'classnames'
import H from 'history'
import PlayCircleOutlineIcon from 'mdi-react/PlayCircleOutlineIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { eventLogger } from '../../../tracking/eventLogger'

interface Props {
    className: string
    location: H.Location
    history: H.History
}

const BASE_VIDEO_URL = 'https://storage.googleapis.com/sourcegraph-assets/video/welcome/video'

// The mp4 videos are 2535x1596 at 20fps.
//
// To convert videos from mp4 (from kazam) to m4v (for iPhone):
//
// ffmpeg -i INPUT_FILE -pix_fmt yuv420p -vf "scale=-2:720:flags=lanczos" -vcodec libx264 -level 3.2 -profile:v main -preset medium -crf 23 -x264-params ref=4 -movflags +faststart OUTPUT_FILE
const VIDEOS: { title: string; hash: string; filename: string }[] = [
    {
        title: 'Go to definition',
        hash: 'go-to-definition',
        filename: 'Welcome-GoToDefinition',
    },
    {
        title: 'Find references',
        hash: 'find-references',
        filename: 'Welcome-FindReferences',
    },
    {
        title: 'Search',
        hash: 'search',
        filename: 'Welcome-Search',
    },
    {
        title: 'On GitHub files',
        hash: 'github-files-integration',
        filename: 'Welcome-GitHub-GoToDefinition',
    },
    {
        title: 'On GitHub PRs',
        hash: 'github-pull-request-integration',
        filename: 'Welcome-GitHub-PR-GoToDefinition',
    },
    {
        title: 'Code coverage',
        hash: 'code-coverage',
        filename: 'Welcome-Coverage',
    },
]

export class WelcomeMainPageDemos extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        const activeTab = this.props.location.hash.replace(/^#/, '') || VIDEOS[0].hash
        const video = VIDEOS.find(v => v.hash === activeTab) || VIDEOS[0]
        return (
            <div className={`welcome-main-page-demos ${this.props.className}`}>
                <ul className="nav nav-pills text-nowrap mb-2 justify-content-md-center">
                    {VIDEOS.map(({ title, hash }, i) => (
                        <li className="nav-item" key={i}>
                            <Link
                                to={{ hash }}
                                className={classnames('welcome-main-page-demos__item nav-link px-2', {
                                    active: activeTab === hash,
                                })}
                                // tslint:disable-next-line:jsx-no-lambda
                                onClick={() => eventLogger.log('WelcomeMainPageDemosVideo', { hash })}
                            >
                                {title}{' '}
                                <PlayCircleOutlineIcon
                                    className={classnames('welcome-main-page-demos__play icon-inline', {
                                        invisible: activeTab !== hash,
                                    })}
                                />
                            </Link>
                        </li>
                    ))}
                </ul>
                <video
                    key={video.hash}
                    autoPlay={true}
                    muted={true}
                    loop={true}
                    playsInline={true}
                    className="welcome-main-page-demos__video w-100 h-auto"
                >
                    <source src={`${BASE_VIDEO_URL}/${video.filename}.mp4`} type="video/mp4" />
                    <source src={`${BASE_VIDEO_URL}/${video.filename}.m4v`} type="video/x-m4v" />
                    Demo video playback is not supported on your browser.
                </video>
            </div>
        )
    }
}
