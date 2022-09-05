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

export const ComputeInsightChart: React.FunctionComponent<
    React.PropsWithChildren<React.SVGProps<SVGSVGElement>>
> = props => (
    <svg
        width="185"
        height="126"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 -15 185 126"
        {...props}
        className={classNames(styles.chart, props.className)}
    >
        <path d="M0 23.7998H137V24.7998H0V23.7998Z" fill="var(--border-color-2)" />
        <path d="M0 0.799805H137V1.7998H0V0.799805Z" fill="var(--border-color-2)" />
        <path d="M0 46.7998H137V47.7998H0V46.7998Z" fill="var(--border-color-2)" />
        <path d="M0 69.7998H137V70.7998H0V69.7998Z" fill="var(--border-color-2)" />
        <path d="M0 92.7998H137V93.7998H0V92.7998Z" fill="var(--border-color-2)" />
        <path d="M164 36.2998H179V37.2998H164V36.2998Z" fill="var(--border-color-2)" />
        <path
            d="M161 36.7998C161 37.9044 160.105 38.7998 159 38.7998C157.895 38.7998 157 37.9044 157 36.7998C157 35.6952 157.895 34.7998 159 34.7998C160.105 34.7998 161 35.6952 161 36.7998Z"
            fill="var(--orange)"
        />
        <path d="M164 45.2998H179V46.2998H164V45.2998Z" fill="var(--border-color-2)" />
        <path
            d="M161 45.7998C161 46.9044 160.105 47.7998 159 47.7998C157.895 47.7998 157 46.9044 157 45.7998C157 44.6952 157.895 43.7998 159 43.7998C160.105 43.7998 161 44.6952 161 45.7998Z"
            fill="var(--pink)"
        />
        <path d="M164 54.2998H179V55.2998H164V54.2998Z" fill="var(--border-color-2)" />
        <path
            d="M161 54.7998C161 55.9044 160.105 56.7998 159 56.7998C157.895 56.7998 157 55.9044 157 54.7998C157 53.6952 157.895 52.7998 159 52.7998C160.105 52.7998 161 53.6952 161 54.7998Z"
            fill="var(--dark-blue)"
        />
        <path d="M164 63.2998H179V64.2998H164V63.2998Z" fill="var(--border-color-2)" />
        <path
            d="M161 63.7998C161 64.9044 160.105 65.7998 159 65.7998C157.895 65.7998 157 64.9044 157 63.7998C157 62.6952 157.895 61.7998 159 61.7998C160.105 61.7998 161 62.6952 161 63.7998Z"
            fill="var(--green)"
        />
        <path d="M16 100H31V101H16V100Z" fill="var(--border-color-2)" />
        <path d="M61.5 100H76.5V101H61.5V100Z" fill="var(--border-color-2)" />
        <path d="M107 100H122V101H107V100Z" fill="var(--border-color-2)" />
        <path
            d="M12 3C12 2.44772 12.4477 2 13 2H14C14.5523 2 15 2.44772 15 3V92C15 92.5523 14.5523 93 14 93H13C12.4477 93 12 92.5523 12 92V3Z"
            fill="var(--dark-blue)"
        />
        <path
            d="M18 18C18 17.4477 18.4477 17 19 17H20C20.5523 17 21 17.4477 21 18V92C21 92.5523 20.5523 93 20 93H19C18.4477 93 18 92.5523 18 92V18Z"
            fill="var(--pink)"
        />
        <path
            d="M24 72C24 71.4477 24.4477 71 25 71H26C26.5523 71 27 71.4477 27 72V92C27 92.5523 26.5523 93 26 93H25C24.4477 93 24 92.5523 24 92V72Z"
            fill="var(--green)"
        />
        <path
            d="M30 81C30 80.4477 30.4477 80 31 80H32C32.5523 80 33 80.4477 33 81V92C33 92.5523 32.5523 93 32 93H31C30.4477 93 30 92.5523 30 92V81Z"
            fill="var(--orange)"
        />
        <path
            d="M58 16C58 15.4477 58.4477 15 59 15H60C60.5523 15 61 15.4477 61 16V92C61 92.5523 60.5523 93 60 93H59C58.4477 93 58 92.5523 58 92V16Z"
            fill="var(--dark-blue)"
        />
        <path
            d="M64 32C64 31.4477 64.4477 31 65 31H66C66.5523 31 67 31.4477 67 32V92C67 92.5523 66.5523 93 66 93H65C64.4477 93 64 92.5523 64 92V32Z"
            fill="var(--pink)"
        />
        <path
            d="M70 58C70 57.4477 70.4477 57 71 57H72C72.5523 57 73 57.4477 73 58V92C73 92.5523 72.5523 93 72 93H71C70.4477 93 70 92.5523 70 92V58Z"
            fill="var(--green)"
        />
        <path
            d="M76 41C76 40.4477 76.4477 40 77 40H78C78.5523 40 79 40.4477 79 41V92C79 92.5523 78.5523 93 78 93H77C76.4477 93 76 92.5523 76 92V41Z"
            fill="var(--orange)"
        />
        <path
            d="M104 53C104 52.4477 104.448 52 105 52H106C106.552 52 107 52.4477 107 53V92C107 92.5523 106.552 93 106 93H105C104.448 93 104 92.5523 104 92V53Z"
            fill="var(--dark-blue)"
        />
        <path
            d="M110 68C110 67.4477 110.448 67 111 67H112C112.552 67 113 67.4477 113 68V92C113 92.5523 112.552 93 112 93H111C110.448 93 110 92.5523 110 92V68Z"
            fill="var(--pink)"
        />
        <path
            d="M116 38C116 37.4477 116.448 37 117 37H118C118.552 37 119 37.4477 119 38V92C119 92.5523 118.552 93 118 93H117C116.448 93 116 92.5523 116 92V38Z"
            fill="var(--green)"
        />
        <path
            d="M122 21C122 20.4477 122.448 20 123 20H124C124.552 20 125 20.4477 125 21V92C125 92.5523 124.552 93 124 93H123C122.448 93 122 92.5523 122 92V21Z"
            fill="var(--orange)"
        />
    </svg>
)
