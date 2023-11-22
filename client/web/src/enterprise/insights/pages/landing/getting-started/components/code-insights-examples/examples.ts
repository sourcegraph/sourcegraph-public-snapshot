import { DATA_SERIES_COLORS } from '../../../../../constants'

import type { CaptureGroupExampleContent, SearchInsightExampleContent } from './types'

interface Datum {
    x: Date
    y: number
}

export const CSS_MODULES_VS_GLOBAL_STYLES_INSIGHT: SearchInsightExampleContent<Datum> = {
    title: 'Migration to CSS modules',
    repositories: 'repo:github.com/wildcard-org/wc-repo',
    series: [
        {
            id: 'series001',
            data: [
                { x: new Date('May 7, 2021'), y: 88 },
                { x: new Date('June 7, 2021'), y: 95 },
                { x: new Date('July 7, 2021'), y: 110 },
                { x: new Date('August 7, 2021'), y: 160 },
                { x: new Date('September 7, 2021'), y: 310 },
                { x: new Date('October 7, 2021'), y: 520 },
                { x: new Date('November 7, 2021'), y: 700 },
            ],
            name: 'CSS Modules',
            query: 'select:file lang:scss file:module.scss patterntype:regexp',
            color: DATA_SERIES_COLORS.GREEN,
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
        {
            id: 'series002',
            data: [
                { x: new Date('May 7, 2021'), y: 410 },
                { x: new Date('June 7, 2021'), y: 410 },
                { x: new Date('July 7, 2021'), y: 315 },
                { x: new Date('August 7, 2021'), y: 180 },
                { x: new Date('September 7, 2021'), y: 90 },
                { x: new Date('October 7, 2021'), y: 45 },
                { x: new Date('November 7, 2021'), y: 10 },
            ],
            name: 'Global CSS',
            query: 'select:file lang:scss -file:module.scss patterntype:regexp',
            color: DATA_SERIES_COLORS.RED,
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
    ],
}

export const ALPINE_VERSIONS_INSIGHT: CaptureGroupExampleContent<Datum> = {
    title: 'Alpine versions over all repos',
    repositories: 'All repositories',
    groupSearch: 'patterntype:regexp FROM\\s+alpine:([\\d\\.]+) file:Dockerfile',
    series: [
        {
            id: 'a',
            data: [
                { x: new Date('May 7, 2021'), y: 100 },
                { x: new Date('June 7, 2021'), y: 90 },
                { x: new Date('July 7, 2021'), y: 85 },
                { x: new Date('August 7, 2021'), y: 85 },
                { x: new Date('September 7, 2021'), y: 70 },
                { x: new Date('October 7, 2021'), y: 50 },
                { x: new Date('November 7, 2021'), y: 35 },
            ],
            name: '3.1',
            color: DATA_SERIES_COLORS.INDIGO,
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
        {
            id: 'b',
            data: [
                { x: new Date('May 7, 2021'), y: 160 },
                { x: new Date('June 7, 2021'), y: 155 },
                { x: new Date('July 7, 2021'), y: 150 },
                { x: new Date('August 7, 2021'), y: 150 },
                { x: new Date('September 7, 2021'), y: 155 },
                { x: new Date('October 7, 2021'), y: 150 },
                { x: new Date('November 7, 2021'), y: 160 },
            ],
            name: '3.5',
            color: DATA_SERIES_COLORS.RED,
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
        {
            id: 'c',
            data: [
                { x: new Date('May 7, 2021'), y: 90 },
                { x: new Date('June 7, 2021'), y: 95 },
                { x: new Date('July 7, 2021'), y: 110 },
                { x: new Date('August 7, 2021'), y: 125 },
                { x: new Date('September 7, 2021'), y: 125 },
                { x: new Date('October 7, 2021'), y: 145 },
                { x: new Date('November 7, 2021'), y: 175 },
            ],
            name: '3.15',
            color: DATA_SERIES_COLORS.GREEN,
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
        {
            id: 'd',
            data: [
                { x: new Date('May 7, 2021'), y: 75 },
                { x: new Date('June 7, 2021'), y: 85 },
                { x: new Date('July 7, 2021'), y: 90 },
                { x: new Date('August 7, 2021'), y: 80 },
                { x: new Date('September 7, 2021'), y: 75 },
                { x: new Date('October 7, 2021'), y: 70 },
                { x: new Date('November 7, 2021'), y: 75 },
            ],
            name: '3.8',
            color: DATA_SERIES_COLORS.GRAPE,
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
        {
            id: 'e',
            data: [
                { x: new Date('May 7, 2021'), y: 85 },
                { x: new Date('June 7, 2021'), y: 80 },
                { x: new Date('July 7, 2021'), y: 60 },
                { x: new Date('August 7, 2021'), y: 50 },
                { x: new Date('September 7, 2021'), y: 45 },
                { x: new Date('October 7, 2021'), y: 35 },
                { x: new Date('November 7, 2021'), y: 45 },
            ],
            name: '3.9',
            color: DATA_SERIES_COLORS.ORANGE,
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
        {
            id: 'f',
            data: [
                { x: new Date('May 7, 2021'), y: 20 },
                { x: new Date('June 7, 2021'), y: 25 },
                { x: new Date('July 7, 2021'), y: 40 },
                { x: new Date('August 7, 2021'), y: 50 },
                { x: new Date('September 7, 2021'), y: 55 },
                { x: new Date('October 7, 2021'), y: 60 },
                { x: new Date('November 7, 2021'), y: 65 },
            ],
            name: '3.9.2',
            color: DATA_SERIES_COLORS.TEAL,
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
        {
            id: 'g',
            name: '3.14',
            data: [
                { x: new Date('May 7, 2021'), y: 150 },
                { x: new Date('June 7, 2021'), y: 155 },
                { x: new Date('July 7, 2021'), y: 165 },
                { x: new Date('August 7, 2021'), y: 165 },
                { x: new Date('September 7, 2021'), y: 160 },
                { x: new Date('October 7, 2021'), y: 155 },
                { x: new Date('November 7, 2021'), y: 145 },
            ],
            color: DATA_SERIES_COLORS.PINK,
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
    ],
}

export const LOG_4_J_INCIDENT_INSIGHT: SearchInsightExampleContent<Datum> = {
    title: 'Log4j incident response',
    repositories: 'repo:github.com/wildcard-org/wc-repo',
    series: [
        {
            id: 'a',
            data: [
                { x: new Date('August 1, 2021'), y: 0 },
                { x: new Date('September 1, 2021'), y: 2 },
                { x: new Date('October 1, 2021'), y: 35 },
                { x: new Date('November 1, 2021'), y: 120 },
                { x: new Date('December 1, 2021'), y: 100 },
                { x: new Date('January 1, 2022'), y: 120 },
                { x: new Date('February 1, 2022'), y: 1500 },
            ],
            name: 'Updated log4j',
            color: DATA_SERIES_COLORS.GREEN,
            query: 'lang:gradle org\\.apache\\.logging\\.log4j[\'"] 2\\.(17)(\\.[0-9]+) patterntype:regexp',
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
        {
            id: 'b',
            name: 'Vulnerable log4j',
            data: [
                { x: new Date('August 1, 2021'), y: 510 },
                { x: new Date('September 1, 2021'), y: 440 },
                { x: new Date('October 1, 2021'), y: 445 },
                { x: new Date('November 1, 2021'), y: 460 },
                { x: new Date('December 1, 2021'), y: 430 },
                { x: new Date('January 1, 2022'), y: 410 },
                { x: new Date('February 1, 2022'), y: 200 },
            ],
            color: DATA_SERIES_COLORS.RED,
            query: 'lang:gradle org\\.apache\\.logging\\.log4j[\'"] 2\\.(0|1|2|3|4|5|6|7|8|9|10|11|12|13|14|15|16)(\\.[0-9]+) patterntype:regexp',
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
    ],
}

export const OPENSSL_PYTHON: SearchInsightExampleContent<Datum> = {
    title: 'Find vulernable OpenSSL versions in the Python Ecosystem',
    repositories: 'All repositories',
    series: [
        {
            id: 'a',
            data: [
                { x: new Date('October 28, 2022'), y: 385 },
                { x: new Date('October 29, 2022'), y: 385 },
                { x: new Date('October 30, 2022'), y: 386 },
                { x: new Date('November 01, 2022'), y: 386 },
                { x: new Date('November 02, 2022'), y: 378 },
                { x: new Date('November 05, 2022'), y: 378 },
                { x: new Date('November 07, 2022'), y: 367 },
            ],
            name: 'pip/pipenv',
            color: DATA_SERIES_COLORS.BLUE,
            query: 'file:requirements.*txt cryptography(s*[=~]=s*(36.|37.|38.0.[0-2])) patternType:regexp archived:no fork:no',
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
    ],
}

export const DEPRECATED_API_USAGE_BY_TEAM: SearchInsightExampleContent<Datum> = {
    title: 'Deprecated API usage by team',
    repositories: 'All repositories',
    series: [
        {
            id: 'a',
            name: 'Cloud',
            data: [
                { x: new Date('August 1, 2021'), y: 165 },
                { x: new Date('September 1, 2021'), y: 180 },
                { x: new Date('October 1, 2021'), y: 125 },
                { x: new Date('November 1, 2021'), y: 80 },
                { x: new Date('December 1, 2021'), y: 120 },
                { x: new Date('January 1, 2022'), y: 140 },
                { x: new Date('February 1, 2022'), y: 100 },
            ],
            color: DATA_SERIES_COLORS.ORANGE,
            query: 'problemAPI file:CloudDirectory',
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
        {
            id: 'b',
            name: 'Core App',
            data: [
                { x: new Date('August 1, 2021'), y: 125 },
                { x: new Date('September 1, 2021'), y: 80 },
                { x: new Date('October 1, 2021'), y: 50 },
                { x: new Date('November 1, 2021'), y: 70 },
                { x: new Date('December 1, 2021'), y: 20 },
                { x: new Date('January 1, 2022'), y: 10 },
                { x: new Date('February 1, 2022'), y: 10 },
            ],
            color: DATA_SERIES_COLORS.CYAN,
            query: 'problemAPI file:CoreDirectory',
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
        {
            id: 'c',
            data: [
                { x: new Date('August 1, 2021'), y: 50 },
                { x: new Date('September 1, 2021'), y: 70 },
                { x: new Date('October 1, 2021'), y: 75 },
                { x: new Date('November 1, 2021'), y: 60 },
                { x: new Date('December 1, 2021'), y: 55 },
                { x: new Date('January 1, 2022'), y: 55 },
                { x: new Date('February 1, 2022'), y: 45 },
            ],
            name: 'Extensibility',
            color: DATA_SERIES_COLORS.GRAPE,
            query: 'problemAPI file:ExtnDirectory',
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
    ],
}

export const LINTER_OVERRIDES: SearchInsightExampleContent<Datum> = {
    title: 'Linter override rules in all repos',
    repositories: 'All repositories',
    series: [
        {
            id: 'a',
            name: 'Linter overrides',
            data: [
                { x: new Date('August 1, 2021'), y: 6800 },
                { x: new Date('September 1, 2021'), y: 12000 },
                { x: new Date('October 1, 2021'), y: 3200 },
                { x: new Date('November 1, 2021'), y: 3600 },
                { x: new Date('December 1, 2021'), y: 3000 },
                { x: new Date('January 1, 2022'), y: 3100 },
                { x: new Date('February 1, 2022'), y: 14500 },
            ],
            color: DATA_SERIES_COLORS.RED,
            query: 'file:\\.eslintignore .\\n patternType:regexp',
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
    ],
}

export const REPOS_WITH_CI_SYSTEM: SearchInsightExampleContent<Datum> = {
    title: '# of repos connected tp the CI system',
    repositories: 'All repositories',
    series: [
        {
            id: 'a',
            data: [
                { x: new Date('August 1, 2021'), y: 60 },
                { x: new Date('September 1, 2021'), y: 60 },
                { x: new Date('October 1, 2021'), y: 120 },
                { x: new Date('November 1, 2021'), y: 80 },
                { x: new Date('December 1, 2021'), y: 200 },
                { x: new Date('January 1, 2022'), y: 325 },
                { x: new Date('February 1, 2022'), y: 480 },
            ],
            name: 'Connected repos',
            color: DATA_SERIES_COLORS.GREEN,
            query: 'file:\\.circleci/config.yml select:repo',
            getXValue: datum => datum.x,
            getYValue: datum => datum.y,
        },
    ],
}
