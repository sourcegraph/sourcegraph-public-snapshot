var sandbox = require("../testSandbox");

var React = require("react");

var CodeListing = require("./CodeListing");
var CodeLineView = require("./CodeLineView");

describe("CodeListing", () => {
	it("should render lines", () => {
		sandbox.renderAndExpect(<CodeListing lines={[{Tokens: ["foo"]}, {}, {Tokens: ["bar"]}]} />).to.eql(
			<div className="code-view-react">
				<table className="line-numbered-code">
					<tbody>
						<CodeLineView lineNumber={1} tokens={["foo"]} key={0} />
						<CodeLineView lineNumber={2} tokens={[]} key={1} />
						<CodeLineView lineNumber={3} tokens={["bar"]} key={2} />
					</tbody>
				</table>
			</div>
		);
	});
});
