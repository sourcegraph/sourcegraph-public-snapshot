// tslint:disable: typedef ordered-imports

import {autotest} from "sourcegraph/util/autotest";

import * as React from "react";

// import * as BuildActions from "sourcegraph/build/BuildActions";
// import {BuildStore} from "sourcegraph/build/BuildStore";
import {BuildContainer} from "sourcegraph/build/BuildContainer";

import testdataInitial from "sourcegraph/build/testdata/BuildContainer-initial.json";
// import testdataAvailable from "sourcegraph/build/testdata/BuildContainer-available.json";
// import testdataBuildSuccess from "sourcegraph/build/testdata/BuildContainer-buildSuccess.json";

describe("BuildContainer", () => {
	it("should initially render empty and want build and tasks", () => {
		autotest(testdataInitial, "sourcegraph/build/testdata/BuildContainer-initial.json",
			<BuildContainer	params={{splat: "r", id: "1"}} />,
			{user: null},
		);
	});

	// TODO(sqs): reenable tests
	/* it("should render tasks", () => {
	   BuildStore.directDispatch(new BuildActions.TasksFetched("aRepo", 123, {BuildTasks: [{ID: 456}]}));
	   autotest(testdataAvailable, "sourcegraph/build/testdata/BuildContainer-available.json",
	   <BuildContainer
	   build={{ID: 123, Repo: "aRepo"}}
	   commit={{ID: "abc"}} />
	   );
	   });

	   it("should render updated build", () => {
	   BuildStore.directDispatch(new BuildActions.BuildFetched("aRepo", 123, {ID: 456, Success: true}));
	   autotest(testdataBuildSuccess, "sourcegraph/build/testdata/BuildContainer-buildSuccess.json",
	   <BuildContainer
	   build={{ID: 123, Repo: "aRepo"}}
	   commit={{ID: "abc"}} />
	   );
	   }); */
});
