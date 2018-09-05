// tslint:disable

import * as React from 'react'
import { render } from 'react-dom'
import CloseIcon from 'mdi-react/CloseIcon'

const IconBase = require('react-icon-base')

export class SourcegraphIcon extends React.Component<any, {}> {
    public render(): JSX.Element {
        return (
            <IconBase viewBox="0 0 40 40" {...this.props}>
                <g fill="none" fillRule="evenodd">
                    <path
                        d="M11.5941935,5.12629921 L20.4929032,36.888189 C21.0909677,39.0226772 23.3477419,40.279685 25.5325806,39.6951181 C27.7190323,39.1105512 29.0051613,36.9064567 28.4067742,34.7722835 L19.5064516,3.00944882 C18.9080645,0.875590551 16.6516129,-0.381732283 14.4667742,0.203149606 C12.2822581,0.786771654 10.9958065,2.99149606 11.5941935,5.12598425 L11.5941935,5.12629921 Z"
                        id="Shape"
                        fill="#F96316"
                    />
                    <path
                        d="M28.0722581,5.00598425 L5.7883871,29.5748031 C4.28516129,31.2314961 4.44225806,33.7647244 6.13741935,35.2327559 C7.83258065,36.7004724 10.4245161,36.5474016 11.9277419,34.8913386 L34.2116129,10.3228346 C35.7148387,8.66614173 35.5577419,6.13385827 33.8625806,4.66551181 C32.1667742,3.19653543 29.5741935,3.34992126 28.0722581,5.00566929 L28.0722581,5.00598425 Z"
                        id="Shape"
                        fill="#B200F8"
                    />
                    <path
                        d="M2.82258065,18.6204724 L34.6019355,28.8866142 C36.7525806,29.5811024 39.0729032,28.4412598 39.7841935,26.3395276 C40.4970968,24.2381102 39.3293548,21.9716535 37.1777419,21.2762205 L5.39935484,11.0110236 C3.24774194,10.3159055 0.928387097,11.455748 0.216774194,13.5574803 C-0.494193548,15.6588976 0.673548387,17.9259843 2.82322581,18.6204724 L2.82258065,18.6204724 Z"
                        id="Shape"
                        fill="#00B4F2"
                    />
                </g>
            </IconBase>
        )
    }
}

export function makeSourcegraphIcon(): HTMLElement {
    const el = document.createElement('span')
    el.className = 'sg-icon'
    render(<SourcegraphIcon />, el)
    return el
}

export function makeCloseIcon(): HTMLElement {
    const el = document.createElement('span')
    el.className = 'sg-icon'
    render(<CloseIcon size={17} />, el)
    return el
}

export class ShowFileTree extends React.Component<any, {}> {
    public render(): JSX.Element {
        return (
            <IconBase viewBox="0 0 12 10" {...this.props}>
                <g id="Page-1" stroke="none" strokeWidth="1" fill="none" fillRule="evenodd">
                    <g id="Vector" fillRule="nonzero" fill="#FFFFFF">
                        <path
                            d="M11,10 L1,10 C0.45,10 0,9.55 0,9 L0,1 C0,0.45 0.45,0 1,0 L11,0 C11.55,0 12,0.45 12,1 L12,9 C12,9.55 11.55,10 11,10 Z M3.5,1.25 C3.5,1.1 3.4,1 3.25,1 L1.25,1 C1.1,1 1,1.1 1,1.25 L1,8.75 C1,8.9 1.1,9 1.25,9 L3.25,9 C3.4,9 3.5,8.9 3.5,8.75 L3.5,1.25 Z M9.4,4.8 L6.9,3.05 C6.775,2.975 6.625,3 6.55,3.1 C6.525,3.15 6.5,3.2 6.5,3.25 L6.5,6.75 C6.5,6.9 6.625,7 6.75,7 C6.8,7 6.85,6.975 6.9,6.95 L9.4,5.2 C9.525,5.125 9.55,4.975 9.475,4.85 C9.45,4.825 9.425,4.8 9.4,4.8 Z"
                            id="path0_fill"
                        />
                    </g>
                </g>
            </IconBase>
        )
    }
}

export class ToggleFileTree extends React.Component<any, {}> {
    public render(): JSX.Element {
        return (
            <IconBase viewBox="0 0 12 10" {...this.props}>
                <g id="Page-1" stroke="none" strokeWidth="1" fill="none" fillRule="evenodd">
                    <g id="ToggleFileTree" fillRule="nonzero" fill="#FFFFFF">
                        <path
                            d="M11,10 L1,10 C0.45,10 0,9.55 0,9 L0,1 C0,0.45 0.45,0 1,0 L11,0 C11.55,0 12,0.45 12,1 L12,9 C12,9.55 11.55,10 11,10 Z M3.5,1.25 C3.5,1.1 3.4,1 3.25,1 L1.25,1 C1.1,1 1,1.1 1,1.25 L1,8.75 C1,8.9 1.1,9 1.25,9 L3.25,9 C3.4,9 3.5,8.9 3.5,8.75 L3.5,1.25 Z M9,3.25 C9,3.1 8.875,3 8.75,3 C8.7,3 8.65,3.025 8.6,3.05 L6.125,4.8 C6,4.875 5.975,5.025 6.05,5.15 C6.075,5.175 6.1,5.2 6.125,5.225 L8.6,6.975 C8.725,7.05 8.875,7.025 8.95,6.925 C8.975,6.875 9,6.825 9,6.775 L9,3.25 Z"
                            id="path0_fill"
                        />
                    </g>
                </g>
            </IconBase>
        )
    }
}
