import classNames from 'classnames'
import React from 'react'

import styles from './MediaCharts.module.scss'

export const ThreeLineChart: React.FunctionComponent<React.SVGProps<SVGSVGElement>> = props => (
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

export const FourLineChart: React.FunctionComponent<React.SVGProps<SVGSVGElement>> = props => (
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

export const LangStatsInsightChart: React.FunctionComponent<React.SVGProps<SVGSVGElement>> = props => (
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

export const SearchBasedInsightChart: React.FunctionComponent<React.SVGProps<SVGSVGElement>> = props => (
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
        <circle cx="8" cy="80" r="2" fill="var(--pink)" />
        <circle cx="30.5" cy="85" r="2" fill="var(--pink)" />
        <circle cx="51.5" cy="77.5" r="2" fill="var(--pink)" />
        <circle cx="74.5" cy="93" r="2" fill="var(--pink)" />
        <circle cx="95.5" cy="93" r="2" fill="var(--pink)" />
        <circle cx="116" cy="107.2" r="2" fill="var(--pink)" />
        <circle cx="139.5" cy="107.2" r="2" fill="var(--pink)" />
        <circle cx="30" cy="32" r="2" fill="var(--dark-blue)" />
        <circle cx="7" cy="61" r="2" fill="var(--dark-blue)" />
        <circle cx="7" cy="47" r="2" fill="var(--green)" />
        <circle cx="32" cy="63" r="2" fill="var(--green)" />
        <circle cx="52.5" cy="53" r="2" fill="var(--green)" />
        <circle cx="75.5" cy="52.5" r="2" fill="var(--green)" />
        <circle cx="97" cy="36.5" r="2" fill="var(--green)" />
        <circle cx="117.5" cy="22.5" r="2" fill="var(--green)" />
        <circle cx="141" cy="22.5" r="2" fill="var(--green)" />
        <path d="m6.5 62.5 23-31h22l22.316-12.735h21.517L118 33.5h22.5" stroke="var(--dark-blue)" strokeWidth="1.5" />
        <path d="m7 79.5 23.212 6L51 77.5 74 93h21.577L116 107h23.5" stroke="var(--pink)" strokeWidth="1.5" />
        <path d="m8 47.5 24.212 15L52 53.5l23-.5 21.577-16L117 23h23.5" stroke="var(--green)" strokeWidth="1.5" />
        <path d="m8 104 21 .5 22.667-16 22.15-8H92l24.5-21 25-10" stroke="var(--orange)" strokeWidth="1.5" />
        <circle cx="8.5" cy="103.5" r="2" fill="var(--orange)" />
        <circle cx="29.5" cy="104" r="2" fill="var(--orange)" />
        <circle cx="52" cy="88" r="2" fill="var(--orange)" />
        <circle cx="74" cy="81" r="2" fill="var(--orange)" />
        <circle cx="92.5" cy="80.5" r="2" fill="var(--orange)" />
        <circle cx="117" cy="59.5" r="2" fill="var(--orange)" />
        <circle cx="141" cy="49.5" r="2" fill="var(--orange)" />
        <circle cx="118.5" cy="33.5" r="2" fill="var(--dark-blue)" />
        <circle cx="141" cy="33.5" r="2" fill="var(--dark-blue)" />
        <circle cx="96" cy="19" r="2" fill="var(--dark-blue)" />
        <circle cx="74" cy="19" r="2" fill="var(--dark-blue)" />
        <circle cx="51.5" cy="31" r="2" fill="var(--dark-blue)" />
    </svg>
)

export const CaptureGroupInsightChart: React.FunctionComponent<React.SVGProps<SVGSVGElement>> = props => (
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
        <rect x="170" y="44.5" width="15" height="1" />
        <circle cx="165" cy="45" r="2" fill="var(--orange)" />
        <rect x="170" y="53.5" width="15" height="1" />
        <circle cx="165" cy="54" r="2" fill="var(--pink)" />
        <rect x="170" y="62.5" width="15" height="1" />
        <circle cx="165" cy="63" r="2" fill="var(--dark-blue)" />
        <rect x="170" y="71.5" width="15" height="1" />
        <circle cx="165" cy="72" r="2" fill="var(--green)" />
        <circle cx="8" cy="19.5" r="2" fill="var(--pink)" />
        <circle cx="30.5" cy="19.5" r="2" fill="var(--pink)" />
        <circle cx="51.5" cy="34" r="2" fill="var(--pink)" />
        <circle cx="74.5" cy="34" r="2" fill="var(--pink)" />
        <circle cx="95.5" cy="45" r="2" fill="var(--pink)" />
        <circle cx="118" cy="45.2" r="2" fill="var(--pink)" />
        <circle cx="141" cy="66.2" r="2" fill="var(--pink)" />
        <circle cx="30" cy="86" r="2" fill="var(--dark-blue)" />
        <circle cx="75.5" cy="99" r="2" fill="var(--green)" />
        <circle cx="97" cy="99" r="2" fill="var(--green)" />
        <circle cx="117.5" cy="86" r="2" fill="var(--green)" />
        <circle cx="141" cy="86" r="2" fill="var(--green)" />
        <path d="M29.5 85.5H51.5L73.8164 64.7647H95.3333L118 23.5H140.5" stroke="var(--dark-blue)" strokeWidth="1.5" />
        <path d="M7 19.5H30.2115L51 34H74L95.5769 45H118L140.5 66" stroke="var(--pink)" strokeWidth="1.5" />
        <path d="M75 99H96.5769L117 86H140.5" stroke="var(--green)" strokeWidth="1.5" />
        <path d="M7 43H30L51.6667 72.5H73.8164L92 91.5H116.5L141.5 104.5" stroke="var(--orange)" strokeWidth="1.5" />
        <circle cx="8" cy="43" r="2" fill="var(--orange)" />
        <circle cx="29.5" cy="43" r="2" fill="var(--orange)" />
        <circle cx="52" cy="72" r="2" fill="var(--orange)" />
        <circle cx="74" cy="72.5" r="2" fill="var(--orange)" />
        <circle cx="92" cy="91.5" r="2" fill="var(--orange)" />
        <circle cx="117" cy="92" r="2" fill="var(--orange)" />
        <circle cx="141" cy="104" r="2" fill="var(--orange)" />
        <circle cx="118.5" cy="23.5" r="2" fill="var(--dark-blue)" />
        <circle cx="141" cy="23.5" r="2" fill="var(--dark-blue)" />
        <circle cx="96" cy="64.5" r="2" fill="var(--dark-blue)" />
        <circle cx="74" cy="65.5" r="2" fill="var(--dark-blue)" />
        <circle cx="51.5" cy="85.5" r="2" fill="var(--dark-blue)" />
        <path
            d="M177.882 104.625a5.261 5.261 0 0 1-.807.065c-.275 0-.541-.024-.808-.065v-2.834l-2.018 2.003a6.703 6.703 0 0 1-1.123-1.123l2.003-2.018h-2.834a5.284 5.284 0 0 1-.065-.808c0-.275.024-.541.065-.808h2.834l-2.003-2.018c.154-.202.315-.404.525-.598.194-.21.396-.371.598-.525l2.018 2.003v-2.834c.267-.04.533-.065.808-.065s.541.024.807.065v2.834l2.019-2.003c.404.315.808.719 1.123 1.123l-2.003 2.018h2.834c.041.267.065.533.065.808s-.024.541-.065.808h-2.834l2.003 2.018a4.462 4.462 0 0 1-.525.598c-.194.21-.396.371-.598.525l-2.019-2.003v2.834Zm-8.882 1.68a1.615 1.615 0 1 1 3.23 0 1.615 1.615 0 0 1-3.23 0Z"
            fill="#A6B6D9"
        />
    </svg>
)
