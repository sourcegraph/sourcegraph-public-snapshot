import * as React from 'react'

interface Props {
    icon: React.ComponentType
    title?: string | JSX.Element
    subtitle?: string | JSX.Element
    cta?: JSX.Element
}

export class HeroPage extends React.Component<Props, {}> {
    public render(): JSX.Element | null {
        return (
            <div className='hero-page'>
                <div className='hero-page__icon'>
                    <this.props.icon />
                </div>
                {
                    this.props.title &&
                    <div className='hero-page__title'>
                        {this.props.title}
                    </div>
                }
                {this.props.subtitle && <div className='hero-page__subtitle'>{this.props.subtitle}</div>}
                {this.props.cta && <div className='hero-page__cta'>{this.props.cta}</div>}
            </div>
        )
    }
}
