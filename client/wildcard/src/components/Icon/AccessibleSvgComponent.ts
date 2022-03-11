interface BaseSVGProps extends React.SVGAttributes<SVGElement> {
    /**
     * Shorthand for width={X} and height={X}
     */
    size?: number
}

interface ScreenReaderIconProps extends BaseSVGProps {
    title: string
}

interface HiddenIconProps extends BaseSVGProps {
    'aria-hidden': true | 'true'
}

export type AccessibleSvgComponent = React.FunctionComponent<ScreenReaderIconProps | HiddenIconProps>
