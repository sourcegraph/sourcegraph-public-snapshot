import chromedriver from "chromedriver";
import webdriver from "selenium-webdriver";

export function delay(time) {
	return new Promise(resolve => setTimeout(resolve, time));
}

let crdvIsStarted = false;
export function startChromeDriver() {
	if (crdvIsStarted) return Promise.resolve();
	chromedriver.start();
	process.on("exit", chromedriver.stop);
	crdvIsStarted = true;
	return delay(1000);
}

export function buildWebDriver(extPath) {
	return new webdriver.Builder()
		.usingServer("http://localhost:9515")
		.withCapabilities({
			chromeOptions: {
				args: [`load-extension=${extPath}`]
			}
		})
		.forBrowser("chrome")
		.build();
}

export function flushLogs(driver) {
	driver.manage().logs().get(webdriver.logging.Type.BROWSER).then((entries) => {
		entries.forEach((entry) => {
			console.log('[%s] %s', entry.level.name, entry.message);
		});
	});
	return true;
}
