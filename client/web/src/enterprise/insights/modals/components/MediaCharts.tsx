import React from 'react'

import classNames from 'classnames'

import styles from './MediaCharts.module.scss'

export const ThreeLineChart: React.FunctionComponent<
    React.PropsWithChildren<React.SVGProps<SVGSVGElement>>
> = props => (
    <svg
        width="169"
        height="158"
        viewBox="0 0 169 158"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
        className={classNames(styles.chart, props.className)}
    >
        <circle cx="60.5" cy="74" r="1.5" fill="var(--red)" />
        <rect x="16" y="45" width="137" height="1" />
        <rect x="16" y="22" width="137" height="1" />
        <rect x="16" y="68" width="137" height="1" />
        <rect x="16" y="91" width="137" height="1" />
        <rect x="16" y="114" width="137" height="1" />
        <rect x="23" y="132.5" width="15" height="1" />
        <circle cx="18" cy="133" r="2" fill="var(--red)" />
        <rect x="57" y="132.5" width="15" height="1" />
        <circle cx="52" cy="133" r="2" fill="var(--blue)" />
        <rect x="91" y="132.5" width="15" height="1" />
        <circle cx="86" cy="133" r="2" fill="var(--green)" />
        <circle cx="17.5" cy="107" r="1.5" fill="var(--green)" />
        <circle cx="40" cy="107" r="1.5" fill="var(--green)" />
        <circle cx="61" cy="92.5" r="1.5" fill="var(--green)" />
        <circle cx="84" cy="91" r="1.5" fill="var(--green)" />
        <circle cx="105" cy="103.5" r="1.5" fill="var(--green)" />
        <circle cx="127.5" cy="103.7" r="1.5" fill="var(--green)" />
        <circle cx="149" cy="103.7" r="1.5" fill="var(--green)" />
        <circle cx="40" cy="78" r="1.5" fill="var(--red)" />
        <circle cx="18.5" cy="95.5" r="1.5" fill="var(--red)" />
        <path
            d="M18 96L40.1498 77.8824L61.6667 74L83.8164 70.7647H105.333H127.483L149 63"
            stroke="var(--red)"
            strokeWidth="1.2"
        />
        <path
            d="M18 107H40.2115L61.7885 92.28L84 91L105.577 103.8H127.154H150"
            stroke="var(--green)"
            strokeWidth="1.2"
        />
        <path
            d="M18 93L40.1498 89.1553L61.6667 76.9806L83.8164 78.9029L105.966 68.0097L127.483 54.5534L149 27"
            stroke="var(--blue)"
            strokeWidth="1.2"
        />
        <circle cx="17.5" cy="93" r="1.5" fill="var(--blue)" />
        <circle cx="40" cy="89" r="1.5" fill="var(--blue)" />
        <circle cx="61.5" cy="77" r="1.5" fill="var(--blue)" />
        <circle cx="83.5" cy="79" r="1.5" fill="var(--blue)" />
        <circle cx="106.5" cy="67.5" r="1.5" fill="var(--blue)" />
        <circle cx="127.5" cy="54.5" r="1.5" fill="var(--blue)" />
        <circle cx="148.5" cy="27.5" r="1.5" fill="var(--blue)" />
        <circle cx="127.5" cy="70.5" r="1.5" fill="var(--red)" />
        <circle cx="106.5" cy="71" r="1.5" fill="var(--red)" />
        <circle cx="148.5" cy="63" r="1.5" fill="var(--red)" />
        <circle cx="83.5" cy="71" r="1.5" fill="var(--red)" />
    </svg>
)

export const FourLineChart: React.FunctionComponent<React.PropsWithChildren<React.SVGProps<SVGSVGElement>>> = props => (
    <svg
        width="169"
        height="158"
        viewBox="0 0 169 158"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
        className={classNames(styles.chart, props.className)}
    >
        <rect x="16" y="45" width="137" height="1" />
        <rect x="16" y="22" width="137" height="1" />
        <rect x="16" y="68" width="137" height="1" />
        <rect x="16" y="91" width="137" height="1" />
        <rect x="16" y="114" width="137" height="1" />
        <rect x="23" y="132.5" width="15" height="1" />
        <circle cx="18" cy="133" r="2" fill="var(--orange)" />
        <rect x="57" y="132.5" width="15" height="1" />
        <circle cx="52" cy="133" r="2" fill="var(--pink)" />
        <rect x="89" y="132.5" width="15" height="1" />
        <circle cx="84" cy="133" r="2" fill="var(--dark-blue)" />
        <rect x="123" y="132.5" width="15" height="1" />
        <circle cx="118" cy="133" r="2" fill="var(--light-green)" />
        <circle cx="17.5" cy="45.5" r="1.5" fill="var(--pink)" />
        <circle cx="40" cy="45.5" r="1.5" fill="var(--pink)" />
        <circle cx="61" cy="85" r="1.5" fill="var(--pink)" />
        <circle cx="84" cy="91" r="1.5" fill="var(--pink)" />
        <circle cx="105" cy="91" r="1.5" fill="var(--pink)" />
        <circle cx="126.5" cy="80.7" r="1.5" fill="var(--pink)" />
        <circle cx="149" cy="80.7" r="1.5" fill="var(--pink)" />
        <circle cx="39.5" cy="91.5" r="1.5" fill="var(--dark-blue)" />
        <circle cx="17.5" cy="113.5" r="1.5" fill="var(--dark-blue)" />
        <circle cx="17.5" cy="100.5" r="1.5" fill="var(--light-green)" />
        <circle cx="41.5" cy="100.5" r="1.5" fill="var(--light-green)" />
        <circle cx="62" cy="114.5" r="1.5" fill="var(--light-green)" />
        <circle cx="85" cy="105" r="1.5" fill="var(--light-green)" />
        <circle cx="106.5" cy="104.5" r="1.5" fill="var(--light-green)" />
        <circle cx="127" cy="68" r="1.5" fill="var(--light-green)" />
        <circle cx="150.5" cy="68" r="1.5" fill="var(--light-green)" />
        <path
            d="M16.5 114.5L39.5 91.5H61.5L83.8164 70.7647H105.333L128 45.5H150.5"
            stroke="var(--dark-blue)"
            strokeWidth="1.2"
        />
        <path d="M17 45.5H40.2115L61 85.5L84 91H105.577L126 81H149.5" stroke="var(--pink)" strokeWidth="1.2" />
        <path
            d="M18 100.5H41.2115L62 114.5L85 105H106.577L127 68H150.5"
            stroke="var(--light-green)"
            strokeWidth="1.2"
        />
        <path
            d="M18 93L39 23.5L61.6667 46.5H83.8164L102 23.5H126.5L151.5 46.5"
            stroke="var(--orange)"
            strokeWidth="1.2"
        />
        <circle cx="17.5" cy="93" r="1.5" fill="var(--orange)" />
        <circle cx="39" cy="23.5" r="1.5" fill="var(--orange)" />
        <circle cx="61.5" cy="46.5" r="1.5" fill="var(--orange)" />
        <circle cx="83.5" cy="46.5" r="1.5" fill="var(--orange)" />
        <circle cx="101.5" cy="24" r="1.5" fill="var(--orange)" />
        <circle cx="126.5" cy="23.5" r="1.5" fill="var(--orange)" />
        <circle cx="150.5" cy="45.5" r="1.5" fill="var(--orange)" />
        <circle cx="128" cy="45.5" r="1.5" fill="var(--dark-blue)" />
        <circle cx="105.5" cy="70.5" r="1.5" fill="var(--dark-blue)" />
        <circle cx="83.5" cy="71" r="1.5" fill="var(--dark-blue)" />
        <circle cx="61" cy="91.5" r="1.5" fill="var(--dark-blue)" />
    </svg>
)

export const LangStatsInsightChart: React.FunctionComponent<
    React.PropsWithChildren<React.SVGProps<SVGSVGElement>>
> = props => (
    <svg
        width="169"
        height="158"
        viewBox="0 0 169 158"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
        className={classNames(styles.chart, props.className)}
    >
        <rect x="16" y="45" width="137" height="1" />
        <rect x="16" y="22" width="137" height="1" />
        <rect x="16" y="68" width="137" height="1" />
        <rect x="16" y="91" width="137" height="1" />
        <rect x="16" y="114" width="137" height="1" />
        <rect x="23" y="132.5" width="15" height="1" />
        <circle cx="18" cy="133" r="2" fill="var(--orange)" />
        <rect x="57" y="132.5" width="15" height="1" />
        <circle cx="52" cy="133" r="2" fill="var(--pink)" />
        <rect x="89" y="132.5" width="15" height="1" />
        <circle cx="84" cy="133" r="2" fill="var(--dark-blue)" />
        <rect x="123" y="132.5" width="15" height="1" />
        <circle cx="118" cy="133" r="2" fill="var(--light-green)" />
        <circle cx="84" cy="69" r="42" fill="var(--light-green)" stroke="var(--body-bg)" />
        <path
            d="M47.6269 48C43.9407 54.3848 42 61.6274 42 69L84 69L47.6269 48Z"
            fill="var(--dark-blue)"
            stroke="var(--body-bg)"
        />
        <path
            d="M84 27C89.5155 27 94.977 28.0864 100.073 30.1971C105.168 32.3078 109.798 35.4015 113.698 39.3015C117.599 43.2016 120.692 47.8316 122.803 52.9273C124.914 58.023 126 63.4845 126 69C126 74.5155 124.914 79.977 122.803 85.0727C120.692 90.1684 117.599 94.7984 113.698 98.6985C109.798 102.599 105.168 105.692 100.073 107.803C94.977 109.914 89.5155 111 84 111L84 69L84 27Z"
            fill="var(--orange)"
            stroke="var(--body-bg)"
        />
        <path
            d="M84 111C78.4845 111 73.023 109.914 67.9273 107.803C62.8316 105.692 58.2016 102.599 54.3015 98.6985C50.4014 94.7984 47.3078 90.1684 45.1971 85.0727C43.0864 79.977 42 74.5155 42 69L84 69L84 111Z"
            fill="var(--pink)"
            stroke="var(--body-bg)"
        />
    </svg>
)

export const SearchBasedInsightChart: React.FunctionComponent<
    React.PropsWithChildren<React.SVGProps<SVGSVGElement>>
> = props => (
    <svg
        width="185"
        height="126"
        viewBox="0 0 185 126"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
        className={classNames(styles.chart, props.className)}
    >
        <path
            fill="var(--border-color-2)"
            d="M6 39h137v1H6zM6 16h137v1H6zM6 62h137v1H6zM6 85h137v1H6zM6 108h137v1H6zM170 44.5h15v1h-15z"
        />

        <circle cx="165" cy="45" r="2" fill="var(--orange)" />
        <path fill="var(--border-color-2)" d="M170 53.5h15v1h-15z" />
        <circle cx="165" cy="54" r="2" fill="var(--pink)" />
        <path fill="var(--border-color-2)" d="M170 62.5h15v1h-15z" />
        <circle cx="165" cy="63" r="2" fill="var(--dark-blue)" />
        <path fill="var(--border-color-2)" d="M170 71.5h15v1h-15z" />
        <circle cx="165" cy="72" r="2" fill="var(--green)" />
        <circle cx="8" cy="109" r="2" fill="var(--pink)" />
        <circle cx="30" cy="109" r="2" fill="var(--pink)" />
        <circle cx="52" cy="109" r="2" fill="var(--pink)" />
        <circle cx="74" cy="103.5" r="2" fill="var(--pink)" />
        <circle cx="95.5" cy="99" r="2" fill="var(--pink)" />
        <circle cx="118" cy="92" r="2" fill="var(--pink)" />
        <circle cx="141.5" cy="100" r="2" fill="var(--pink)" />
        <circle cx="30" cy="30" r="2" fill="var(--dark-blue)" />
        <circle cx="8" cy="17" r="2" fill="var(--dark-blue)" />
        <circle cx="8" cy="79" r="2" fill="var(--green)" />
        <circle cx="30" cy="67" r="2" fill="var(--green)" />
        <circle cx="52" cy="52.5" r="2" fill="var(--green)" />
        <circle cx="74" cy="52.5" r="2" fill="var(--green)" />
        <circle cx="96" cy="36.5" r="2" fill="var(--green)" />
        <circle cx="118" cy="22.5" r="2" fill="var(--green)" />
        <circle cx="141" cy="22.5" r="2" fill="var(--green)" />
        <path
            d="m8 17 21.5 12.5 22 7 22.316 53 21.517-72.735L118 31.5h22.5"
            stroke="var(--dark-blue)"
            strokeWidth="1.5"
        />
        <path d="M6.5 109H52l22-5.5 20.5-4L118 92l23 7.5" stroke="var(--pink)" strokeWidth="1.5" />
        <path d="m7 79.5 23.212-12L52 53h22l22-16.5 22-14h22" stroke="var(--green)" strokeWidth="1.5" />
        <path d="m8 100 21-5 22.667-8.5 22.15-22L96 81.5l21.5-8 24-24" stroke="var(--orange)" strokeWidth="1.5" />
        <circle cx="8" cy="100" r="2" fill="var(--orange)" />
        <circle cx="29.5" cy="95" r="2" fill="var(--orange)" />
        <circle cx="52" cy="86" r="2" fill="var(--orange)" />
        <circle cx="74" cy="65" r="2" fill="var(--orange)" />
        <circle cx="96" cy="81.5" r="2" fill="var(--orange)" />
        <circle cx="118" cy="73.5" r="2" fill="var(--orange)" />
        <circle cx="141" cy="50" r="2" fill="var(--orange)" />
        <circle cx="118" cy="32" r="2" fill="var(--dark-blue)" />
        <circle cx="141" cy="32" r="2" fill="var(--dark-blue)" />
        <circle cx="96" cy="17" r="2" fill="var(--dark-blue)" />
        <circle cx="74" cy="90" r="2" fill="var(--dark-blue)" />
        <circle cx="52" cy="37.5" r="2" fill="var(--dark-blue)" />
    </svg>
)

export const CaptureGroupInsightChart: React.FunctionComponent<
    React.PropsWithChildren<React.SVGProps<SVGSVGElement>>
> = props => (
    <svg
        width="185"
        height="126"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 185 126"
        {...props}
        className={classNames(styles.chart, props.className)}
    >
        <path
            fill="var(--border-color-2)"
            d="M6 39h137v1H6zM6 16h137v1H6zM6 62h137v1H6zM6 85h137v1H6zM6 108h137v1H6zM170 44.5h15v1h-15z"
        />
        <circle cx="165" cy="45" r="2" fill="var(--orange)" />
        <path fill="var(--border-color-2)" d="M170 53.5h15v1h-15z" />
        <circle cx="165" cy="54" r="2" fill="var(--pink)" />
        <path fill="var(--border-color-2)" d="M170 62.5h15v1h-15z" />
        <circle cx="165" cy="63" r="2" fill="var(--dark-blue)" />
        <path fill="var(--border-color-2)" d="M170 71.5h15v1h-15z" />
        <circle cx="165" cy="72" r="2" fill="var(--green)" />
        <circle cx="8" cy="57.5" r="2" fill="var(--pink)" />
        <circle cx="30" cy="57.5" r="2" fill="var(--pink)" />
        <circle cx="52" cy="39" r="2" fill="var(--pink)" />
        <circle cx="74" cy="39" r="2" fill="var(--pink)" />
        <circle cx="96" cy="39" r="2" fill="var(--pink)" />
        <circle cx="118" cy="16.2" r="2" fill="var(--pink)" />
        <circle cx="141" cy="16.2" r="2" fill="var(--pink)" />
        <circle cx="30" cy="85.5" r="2" fill="var(--dark-blue)" />
        <circle cx="8" cy="85.5" r="2" fill="var(--dark-blue)" />
        <circle cx="8" cy="108.5" r="2" fill="var(--green)" />
        <circle cx="30" cy="108.5" r="2" fill="var(--green)" />
        <circle cx="52" cy="108.5" r="2" fill="var(--green)" />
        <circle cx="74" cy="108.5" r="2" fill="var(--green)" />
        <circle cx="96" cy="108.5" r="2" fill="var(--green)" />
        <circle cx="118" cy="85" r="2" fill="var(--green)" />
        <circle cx="141" cy="85" r="2" fill="var(--green)" />
        <path d="M8 85.5h44.5l21.316-14.735h22.517L118 39.5h22.5" stroke="var(--dark-blue)" strokeWidth="1.5" />
        <path d="M7 57.5h23.212L52 39h43.577L118 16h21.5" stroke="var(--pink)" strokeWidth="1.5" />
        <path d="M8 108.5h87.577L118 85h22.5" stroke="var(--green)" strokeWidth="1.5" />
        <path d="M8 39h22l21.667 10.5h22.15L96 59.5h45.5" stroke="var(--orange)" strokeWidth="1.5" />
        <circle cx="8" cy="39" r="2" fill="var(--orange)" />
        <circle cx="30" cy="39" r="2" fill="var(--orange)" />
        <circle cx="52" cy="49.5" r="2" fill="var(--orange)" />
        <circle cx="74" cy="49.5" r="2" fill="var(--orange)" />
        <circle cx="96" cy="59.5" r="2" fill="var(--orange)" />
        <circle cx="118" cy="59.5" r="2" fill="var(--orange)" />
        <circle cx="141" cy="59.5" r="2" fill="var(--orange)" />
        <circle cx="141" cy="39.5" r="2" fill="var(--dark-blue)" />
        <circle cx="118" cy="40" r="2" fill="var(--dark-blue)" />
        <circle cx="96" cy="71" r="2" fill="var(--dark-blue)" />
        <circle cx="74" cy="71" r="2" fill="var(--dark-blue)" />
        <circle cx="52" cy="85.5" r="2" fill="var(--dark-blue)" />
        <path
            d="M176.882 105.625a5.261 5.261 0 0 1-.807.065c-.275 0-.541-.024-.808-.065v-2.834l-2.018 2.003a6.703 6.703 0 0 1-1.123-1.123l2.003-2.018h-2.834a5.284 5.284 0 0 1-.065-.808c0-.275.024-.541.065-.808h2.834l-2.003-2.018c.154-.202.315-.404.525-.598.194-.21.396-.371.598-.525l2.018 2.003v-2.834c.267-.04.533-.065.808-.065s.541.024.807.065v2.834l2.019-2.003c.404.315.808.719 1.123 1.123l-2.003 2.018h2.834c.041.267.065.533.065.808s-.024.541-.065.808h-2.834l2.003 2.018a4.462 4.462 0 0 1-.525.598c-.194.21-.396.371-.598.525l-2.019-2.003v2.834Zm-8.882 1.68a1.615 1.615 0 1 1 3.23 0 1.615 1.615 0 0 1-3.23 0Z"
            fill="#A6B6D9"
        />
    </svg>
)
