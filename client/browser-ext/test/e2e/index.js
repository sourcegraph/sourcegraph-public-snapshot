import path from "path";
import webdriver from "selenium-webdriver";
import {expect} from "chai";
import {delay, startChromeDriver, buildWebDriver} from "../utils";

const logging = webdriver.logging;

describe("inject app", function test() {
	let driver;
	this.timeout(30000); // whole test must complete in 30s

	before(async () => {
		const prefs = new logging.Preferences();
		prefs.setLevel(logging.Type.BROWSER, logging.Level.DEBUG);
		const caps = webdriver.Capabilities.chrome();
		caps.setLoggingPrefs(prefs);

		await startChromeDriver();
		const extPath = path.resolve("build");
		driver = buildWebDriver(extPath);

		driver.manage().timeouts().implicitlyWait(5000);

		await driver.get("https://github.com/gorilla/mux");
	});

	after(async () => driver.quit());

	describe("#sourcegraph-app-bootstrap", function test() {
		it("should get added to the document", async () => {
			const numElems = await driver.wait(() => driver.findElements(webdriver.By.id("sourcegraph-app-bootstrap")))
			expect(numElems.length).to.eql(1);
		});

		it("should be hidden", async () => {
			const display = await driver.wait(() => driver.findElement(webdriver.By.id("sourcegraph-app-bootstrap"))
				.then((elem) => elem.getCssValue("display"))
			);
			expect(display).to.eql("none");
		});
	});

	describe("#sourcegraph-app-background", function test() {
		it("should get added to the document", async () => {
			const numElems = await driver.wait(() => driver.findElements(webdriver.By.id("sourcegraph-app-background")));
			expect(numElems.length).to.eql(1);
		});
	});

	describe("BlobAnnotator", function test() {

		describe("blob view", function test() {
			before(async () => await driver.navigate().to("https://github.com/gorilla/mux/blob/757bef944d0f21880861c2dd9c871ca543023cba/mux.go"));

			it("should inject BlobAnnotator", async () => {
				const numElems = await driver.wait(() => driver.findElements(webdriver.By.className("sourcegraph-app-annotator")));
				expect(numElems.length).to.eql(1)
			});

			it("should provide a hover tooltip", async () => {
				await driver.wait(() => driver.findElement(webdriver.By.id("text-node-297-6"))
					.then((elem) => driver.executeScript("if(document.createEvent){var evObj = document.createEvent('MouseEvents');evObj.initEvent('mouseover',true, false); arguments[0].dispatchEvent(evObj);} else if(document.createEventObject) { arguments[0].fireEvent('onmouseover');}", elem))
					.then(() => driver.findElement(webdriver.By.className("sg-popover")))
				);
			});

			it("should provide a jump to def", async () => {
				const jumpToUrl = await driver.wait(() => driver.findElement(webdriver.By.id("text-node-wrapper-297"))
					.then((elem) => elem.findElement(webdriver.By.className("sg-clickable")))
					.then((elem) => elem.click())
					.then(() => driver.sleep(500)) // wait for page navigation
					.then(() => driver.getCurrentUrl())
				);
				expect(jumpToUrl).to.eql("https://github.com/gorilla/mux/blob/757bef944d0f21880861c2dd9c871ca543023cba/mux.go#L17");
			});
		});

		describe("pull request view", function test() {
			before(async () => await driver.navigate().to("https://github.com/gorilla/mux/pull/205/files"));

			it("should inject a BlobAnnotator per file", async () => {
				const numElems = await driver.wait(() => driver.findElements(webdriver.By.className("sourcegraph-app-annotator")));
				expect(numElems.length).to.eql(2)
			});

			describe("code addition", function test() {
				it("should provide a hover tooltip for an addition", async () => {
					await driver.wait(() => driver.findElement(webdriver.By.id("text-node-272-5"))
						.then((elem) => driver.executeScript("if(document.createEvent){var evObj = document.createEvent('MouseEvents');evObj.initEvent('mouseover',true, false); arguments[0].dispatchEvent(evObj);} else if(document.createEventObject) { arguments[0].fireEvent('onmouseover');}", elem))
						.then(() => driver.findElement(webdriver.By.className("sg-popover")))
					);
				});

				it("should provide a jump to def", async () => {
					const jumpToUrl = await driver.wait(() => driver.findElement(webdriver.By.id("text-node-wrapper-272"))
						.then((elem) => elem.findElement(webdriver.By.className("sg-clickable")))
						.then((elem) => elem.click())
						.then(() => driver.sleep(500)) // wait for page navigation
						.then(() => driver.getCurrentUrl())
					);
					expect(jumpToUrl).to.eql("https://github.com/captncraig/mux/blob/acfc892941192f90aadd4f452a295bf39fc5f7ed/mux.go#L17");
				});
			});
		});

	});
});
