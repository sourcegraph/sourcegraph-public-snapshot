import shallowRender from "./shallowRender";
import expect from "expect.js";

import React from "react";
import Draggable from "react-draggable";

import Dispatcher from "./Dispatcher";
import * as DefActions from "./DefActions";
import DefPopup from "./DefPopup";

describe("DefPopup", () => {
	it("should render definition data", () => {
		shallowRender(
			<DefPopup def={{URL: "someURL", QualifiedName: "someName", Data: {DocHTML: "someDoc"}}} />
		).compare(
			<Draggable handle="header.toolbar">
				<div className="token-details">
					<div className="body">
						<header className="toolbar">
							<a className="btn btn-toolbar btn-default go-to-def" href="someURL">Go to definition</a>
							<a className="close top-action">×</a>
						</header>
						<section className="docHTML">
							<div className="header">
								<h1 className="qualified-name" dangerouslySetInnerHTML="someName" />
							</div>
							<section className="doc" dangerouslySetInnerHTML="someDoc" />
						</section>
					</div>
				</div>
			</Draggable>
		);
	});

	it("should unselect definition on close", () => {
		expect(Dispatcher.catchDispatched(() => {
			shallowRender(<DefPopup def={{}} />).querySelector(".close").props.onClick();
		})).to.eql([new DefActions.SelectDef(null)]);
	});

	it("should go to definition on button click", () => {
		let defaultPrevented = false;
		expect(Dispatcher.catchDispatched(() => {
			shallowRender(<DefPopup def={{URL: "someURL"}} />).querySelector(".go-to-def").props.onClick({preventDefault() { defaultPrevented = true; }});
		})).to.eql([new DefActions.GoToDef("someURL")]);
		expect(defaultPrevented).to.be(true);
	});

});
