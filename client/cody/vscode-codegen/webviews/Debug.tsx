import './Debug.css'

interface DebugProps {
	debugLog: string[]
}

export const Debug: React.FunctionComponent<React.PropsWithChildren<DebugProps>> = ({ debugLog }) => (
	<div className="inner-container">
		<div className="non-transcript-container">
			<div className="debug-container" data-tab-target="debug">
				{debugLog &&
					debugLog.map((log, i) => (
						<div key={`log-${i}`} className="debug-message">
							{log}
						</div>
					))}
			</div>
		</div>
	</div>
)
