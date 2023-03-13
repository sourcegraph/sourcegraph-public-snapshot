import React from 'react'

export const CodySvg = React.memo(() => (
	<div className="header-logo">
		<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24">
			<path
				d="M9 15a1 1 0 0 1-1-1v-2a1 1 0 0 1 2 0v2a1 1 0 0 1-1 1ZM15 15a1 1 0 0 1-1-1v-2a1 1 0 0 1 2 0v2a1 1 0 0 1-1 1ZM6 8a1 1 0 0 1-.71-.29l-3-3a1 1 0 0 1 1.42-1.42l3 3a1 1 0 0 1 0 1.42A1 1 0 0 1 6 8ZM18 8a1 1 0 0 1-.71-.29 1 1 0 0 1 0-1.42l3-3a1 1 0 1 1 1.42 1.42l-3 3A1 1 0 0 1 18 8Z"
				fill="var(--vscode-icon-foreground)"
			/>
			<path
				d="M21 20H3a1 1 0 0 1-1-1v-4.5a10 10 0 0 1 20 0V19a1 1 0 0 1-1 1ZM4 18h16v-3.5a8 8 0 0 0-16 0Z"
				fill="var(--vscode-icon-foreground)"
			/>
		</svg>
	</div>
))

export const ResetSvg = React.memo(() => (
	<div className="header-container-right">
		<div className="reset-conversation" title="Start a new conversation with Cody">
			<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="var(--vscode-icon-foreground)">
				<path d="M12 16c1.671 0 3-1.331 3-3s-1.329-3-3-3-3 1.331-3 3 1.329 3 3 3z" />
				<path d="M20.817 11.186a8.94 8.94 0 0 0-1.355-3.219 9.053 9.053 0 0 0-2.43-2.43 8.95 8.95 0 0 0-3.219-1.355 9.028 9.028 0 0 0-1.838-.18V2L8 5l3.975 3V6.002c.484-.002.968.044 1.435.14a6.961 6.961 0 0 1 2.502 1.053 7.005 7.005 0 0 1 1.892 1.892A6.967 6.967 0 0 1 19 13a7.032 7.032 0 0 1-.55 2.725 7.11 7.11 0 0 1-.644 1.188 7.2 7.2 0 0 1-.858 1.039 7.028 7.028 0 0 1-3.536 1.907 7.13 7.13 0 0 1-2.822 0 6.961 6.961 0 0 1-2.503-1.054 7.002 7.002 0 0 1-1.89-1.89A6.996 6.996 0 0 1 5 13H3a9.02 9.02 0 0 0 1.539 5.034 9.096 9.096 0 0 0 2.428 2.428A8.95 8.95 0 0 0 12 22a9.09 9.09 0 0 0 1.814-.183 9.014 9.014 0 0 0 3.218-1.355 8.886 8.886 0 0 0 1.331-1.099 9.228 9.228 0 0 0 1.1-1.332A8.952 8.952 0 0 0 21 13a9.09 9.09 0 0 0-.183-1.814z" />
			</svg>
		</div>
	</div>
))

export const SubmitSvg = React.memo(() => (
	<div className="submit-container">
		<svg
			xmlns="http://www.w3.org/2000/svg"
			width="24"
			height="24"
			viewBox="0 0 24 24"
			fill="var(--vscode-icon-foreground)"
		>
			<path d="m21.426 11.095-17-8A1 1 0 0 0 3.03 4.242l1.212 4.849L12 12l-7.758 2.909-1.212 4.849a.998.998 0 0 0 1.396 1.147l17-8a1 1 0 0 0 0-1.81z" />
		</svg>
	</div>
))

export const RightArrowSvg = React.memo(() => (
	<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="var(--vscode-icon-foreground)">
		<path d="m9 19 8-7-8-7z" />
	</svg>
))

export const DownArrowSvg = React.memo(() => (
	<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="var(--vscode-icon-foreground)">
		<path d="m11.998 17 7-8h-14z" />
	</svg>
))
