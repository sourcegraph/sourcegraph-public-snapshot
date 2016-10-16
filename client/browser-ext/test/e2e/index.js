import path from "path";
import webdriver from "selenium-webdriver";
import {expect} from "chai";
import {delay, startChromeDriver, buildWebDriver} from "../utils";

describe("inject app", function test() {
	let driver;
	this.timeout(10000); // whole test must complete in 10s

	before(async () => {
		await startChromeDriver();
		const extPath = path.resolve("build");
		driver = buildWebDriver(extPath);
		await driver.get("https://github.com/gorilla/mux");
	});

	after(async () => driver.quit());

	it("should open github.com/gorilla/mux", async () => {
		const title = await driver.getTitle();
		expect(title).to.equal("GitHub - gorilla/mux: A powerful URL router and dispatcher for golang.");
	});

	describe("#sourcegraph-app-bootstrap", function test() {

		it("should get added to the document", async () => {
			await driver.wait(
				() => driver.findElements(webdriver.By.id("sourcegraph-app-bootstrap"))
					.then((elems) => elems.length === 1),
				1000
			);
		});

		it("should be hidden", async () => {
			await driver.wait(
				() => driver.findElement(webdriver.By.id("sourcegraph-app-bootstrap"))
					.then((elem) => elem.getCssValue("display"))
					.then((val) => val === "none"),
				1000
			);
		});
	});

	describe("#sourcegraph-app-background", function test() {
		it("should get added to the document", async () => {
			await driver.wait(
				() => driver.findElements(webdriver.By.id("sourcegraph-app-background"))
					.then((elems) => elems.length === 1),
				1000
			);
		});
	});
});
