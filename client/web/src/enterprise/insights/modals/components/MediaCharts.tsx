import classname from 'classnames'
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
        className={classname(styles.chart, props.className)}
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
        className={classname(styles.chart, props.className)}
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

export const PieChart: React.FunctionComponent<React.SVGProps<SVGSVGElement>> = props => (
    <svg
        width="169"
        height="158"
        viewBox="0 0 169 158"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
        className={classname(styles.chart, props.className)}
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
