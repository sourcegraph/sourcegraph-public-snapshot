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
        <circle cx="60.5" cy="74" r="1.5" />
        <rect x="16" y="45" width="137" height="1" />
        <rect x="16" y="22" width="137" height="1" />
        <rect x="16" y="68" width="137" height="1" />
        <rect x="16" y="91" width="137" height="1" />
        <rect x="16" y="114" width="137" height="1" />
        <rect x="23" y="132.5" width="15" height="1" />
        <circle cx="18" cy="133" r="2" fill="#DD4E47" />
        <rect x="57" y="132.5" width="15" height="1" />
        <circle cx="52" cy="133" r="2" fill="#3C7ED0" />
        <rect x="91" y="132.5" width="15" height="1" />
        <circle cx="86" cy="133" r="2" fill="#4496AA" />
        <circle cx="17.5" cy="107" r="1.5" fill="#4496AA" />
        <circle cx="40" cy="107" r="1.5" fill="#4496AA" />
        <circle cx="61" cy="92.5" r="1.5" fill="#4496AA" />
        <circle cx="84" cy="91" r="1.5" fill="#4496AA" />
        <circle cx="105" cy="103.5" r="1.5" fill="#4496AA" />
        <circle cx="127.5" cy="103.7" r="1.5" fill="#4496AA" />
        <circle cx="149" cy="103.7" r="1.5" fill="#4496AA" />
        <circle cx="40" cy="78" r="1.5" fill="#DD4E47" />
        <circle cx="18.5" cy="95.5" r="1.5" fill="#DD4E47" />
        <path
            d="M18 96L40.1498 77.8824L61.6667 74L83.8164 70.7647H105.333H127.483L149 63"
            stroke="#DD4E47"
            strokeWidth="1.2"
        />
        <path d="M18 107H40.2115L61.7885 92.28L84 91L105.577 103.8H127.154H150" stroke="#4496AA" strokeWidth="1.2" />
        <path
            d="M18 93L40.1498 89.1553L61.6667 76.9806L83.8164 78.9029L105.966 68.0097L127.483 54.5534L149 27"
            stroke="#3C7ED0"
            strokeWidth="1.2"
        />
        <circle cx="17.5" cy="93" r="1.5" fill="#3C7ED0" />
        <circle cx="40" cy="89" r="1.5" fill="#3C7ED0" />
        <circle cx="61.5" cy="77" r="1.5" fill="#3C7ED0" />
        <circle cx="83.5" cy="79" r="1.5" fill="#3C7ED0" />
        <circle cx="106.5" cy="67.5" r="1.5" fill="#3C7ED0" />
        <circle cx="127.5" cy="54.5" r="1.5" fill="#3C7ED0" />
        <circle cx="148.5" cy="27.5" r="1.5" fill="#3C7ED0" />
        <circle cx="127.5" cy="70.5" r="1.5" fill="#DD4E47" />
        <circle cx="106.5" cy="71" r="1.5" fill="#DD4E47" />
        <circle cx="148.5" cy="63" r="1.5" fill="#DD4E47" />
        <circle cx="83.5" cy="71" r="1.5" fill="#DD4E47" />
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
        <circle cx="18" cy="133" r="2" fill="#E77334" />
        <rect x="57" y="132.5" width="15" height="1" />
        <circle cx="52" cy="133" r="2" fill="#C5436C" />
        <rect x="89" y="132.5" width="15" height="1" />
        <circle cx="84" cy="133" r="2" fill="#4865E3" />
        <rect x="123" y="132.5" width="15" height="1" />
        <circle cx="118" cy="133" r="2" fill="#4BA37B" />
        <circle cx="17.5" cy="45.5" r="1.5" fill="#C5436C" />
        <circle cx="40" cy="45.5" r="1.5" fill="#C5436C" />
        <circle cx="61" cy="85" r="1.5" fill="#C5436C" />
        <circle cx="84" cy="91" r="1.5" fill="#C5436C" />
        <circle cx="105" cy="91" r="1.5" fill="#C5436C" />
        <circle cx="126.5" cy="80.7" r="1.5" fill="#C5436C" />
        <circle cx="149" cy="80.7" r="1.5" fill="#C5436C" />
        <circle cx="39.5" cy="91.5" r="1.5" fill="#4865E3" />
        <circle cx="17.5" cy="113.5" r="1.5" fill="#4865E3" />
        <circle cx="17.5" cy="100.5" r="1.5" fill="#4BA37B" />
        <circle cx="41.5" cy="100.5" r="1.5" fill="#4BA37B" />
        <circle cx="62" cy="114.5" r="1.5" fill="#4BA37B" />
        <circle cx="85" cy="105" r="1.5" fill="#4BA37B" />
        <circle cx="106.5" cy="104.5" r="1.5" fill="#4BA37B" />
        <circle cx="127" cy="68" r="1.5" fill="#4BA37B" />
        <circle cx="150.5" cy="68" r="1.5" fill="#4BA37B" />
        <path
            d="M16.5 114.5L39.5 91.5H61.5L83.8164 70.7647H105.333L128 45.5H150.5"
            stroke="#4865E3"
            strokeWidth="1.2"
        />
        <path d="M17 45.5H40.2115L61 85.5L84 91H105.577L126 81H149.5" stroke="#C5436C" strokeWidth="1.2" />
        <path d="M18 100.5H41.2115L62 114.5L85 105H106.577L127 68H150.5" stroke="#4BA37B" strokeWidth="1.2" />
        <path d="M18 93L39 23.5L61.6667 46.5H83.8164L102 23.5H126.5L151.5 46.5" stroke="#E77334" strokeWidth="1.2" />
        <circle cx="17.5" cy="93" r="1.5" fill="#E77334" />
        <circle cx="39" cy="23.5" r="1.5" fill="#E77334" />
        <circle cx="61.5" cy="46.5" r="1.5" fill="#E77334" />
        <circle cx="83.5" cy="46.5" r="1.5" fill="#E77334" />
        <circle cx="101.5" cy="24" r="1.5" fill="#E77334" />
        <circle cx="126.5" cy="23.5" r="1.5" fill="#E77334" />
        <circle cx="150.5" cy="45.5" r="1.5" fill="#E77334" />
        <circle cx="128" cy="45.5" r="1.5" fill="#4865E3" />
        <circle cx="105.5" cy="70.5" r="1.5" fill="#4865E3" />
        <circle cx="83.5" cy="71" r="1.5" fill="#4865E3" />
        <circle cx="61" cy="91.5" r="1.5" fill="#4865E3" />
    </svg>
)

export const TwoLineChart: React.FunctionComponent<React.SVGProps<SVGSVGElement>> = props => (
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
        <circle cx="18" cy="133" r="2" fill="#4496AA" />
        <rect x="57" y="132.5" width="15" height="1" />
        <circle cx="52" cy="133" r="2" fill="#3C7ED0" />
        <circle cx="17.5" cy="85" r="1.5" fill="#4496AA" />
        <circle cx="40" cy="84" r="1.5" fill="#4496AA" />
        <circle cx="61.5" cy="72.5" r="1.5" fill="#4496AA" />
        <circle cx="84" cy="67" r="1.5" fill="#4496AA" />
        <circle cx="106" cy="50" r="1.5" fill="#4496AA" />
        <circle cx="127.5" cy="31.7" r="1.5" fill="#4496AA" />
        <circle cx="151" cy="113.7" r="1.5" fill="#3C7ED0" />
        <path
            d="M18 85L40.2115 84L61.7885 72.28L84 67L105.577 50.8L127.154 31.8L150 22.8"
            stroke="#4496AA"
            strokeWidth="1.2"
        />
        <path
            d="M18 93L40.1498 96.1553L61.6667 104.981L83.8164 104.903L105.966 114.01L127.483 110.553L151 114"
            stroke="#3C7ED0"
            strokeWidth="1.2"
        />
        <circle cx="17.5" cy="93" r="1.5" fill="#3C7ED0" />
        <circle cx="40" cy="96" r="1.5" fill="#3C7ED0" />
        <circle cx="61.5" cy="105" r="1.5" fill="#3C7ED0" />
        <circle cx="83.5" cy="105" r="1.5" fill="#3C7ED0" />
        <circle cx="106" cy="114" r="1.5" fill="#3C7ED0" />
        <circle cx="127.5" cy="110.5" r="1.5" fill="#3C7ED0" />
        <circle cx="149.5" cy="23" r="1.5" fill="#4496AA" />
    </svg>
)
