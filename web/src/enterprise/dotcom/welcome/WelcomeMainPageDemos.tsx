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
//
// To upload files:
//
// gsutil cp -a public-read -r INPUT_FILES gs://sourcegraph-assets/video/welcome/video/
const VIDEOS: { title: string; hash: string; filename: string }[] = [
    {
        title: 'Code navigation',
        hash: 'code-navigation',
        filename: 'Welcome-CodeNavigation',
    },
    {
        title: 'Code search',
        hash: 'code-search',
        filename: 'Welcome-Search',
    },
    {
        title: 'GitHub integration',
        hash: 'github-integration',
        filename: 'Welcome-GitHub',
    },
]

export class WelcomeMainPageDemos extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        const activeTab = this.props.location.hash.replace(/^#/, '') || VIDEOS[0].hash
        const video = VIDEOS.find(v => v.hash === activeTab) || VIDEOS[0]
        return (
            <div className={`welcome-main-page-demos ${this.props.className}`}>
                <ul className="nav nav-pills text-nowrap justify-content-md-center">
                    <li className="nav-item disabled text-muted d-flex align-items-center mr-2 mb-2">
                        <PlayCircleOutlineIcon className="mr-2" /> Demos:
                    </li>
                    {VIDEOS.map(({ title, hash }, i) => (
                        <li className="nav-item" key={i}>
                            <Link
                                to={{ hash }}
                                className={classnames('welcome-main-page-demos__item nav-link border mx-1 mb-2', {
                                    active: activeTab === hash,
                                })}
                                // tslint:disable-next-line:jsx-no-lambda
                                onClick={() => eventLogger.log('WelcomeMainPageDemosVideo', { hash })}
                            >
                                {title}
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
