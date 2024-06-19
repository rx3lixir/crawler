const puppeteer = require("puppeteer");
const winston = require("winston");

const logger = winston.createLogger({
  level: "info",
  format: winston.format.combine(
    winston.format.colorize(),
    winston.format.timestamp(),
    winston.format.printf(
      ({ timestamp, level, message }) => `${timestamp} [${level}]: ${message}`,
    ),
  ),
  transports: [new winston.transports.Console()],
});

let browser;
let pagePool = [];

async function initializeBrowser() {
  if (!browser) {
    browser = await puppeteer.launch();
    logger.info("Browser initialized");
  }
}

async function getPage() {
  if (pagePool.length > 0) {
    return pagePool.pop();
  }
  return await browser.newPage();
}

async function releasePage(page) {
  pagePool.push(page);
}

async function scrapePage(url, selector) {
  const page = await getPage();

  try {
    logger.info(`Navigating to ${url}`);
    await page.goto(url, { waitUntil: "networkidle2", timeout: 60000 });

    logger.info(`Waiting for selector ${selector}`);
    await page.waitForSelector(selector, { timeout: 60000 });

    const elementExists = (await page.$(selector)) !== null;
    logger.info(
      `Element with selector ${selector} ${elementExists ? "found" : "not found"} on ${url}`,
    );

    const html = await page.content();
    logger.info("Page content fetched");
    return html;
  } catch (err) {
    logger.error(`Error: ${err.message}`);
    return null;
  } finally {
    await releasePage(page);
  }
}

async function closeBrowser() {
  if (browser) {
    await browser.close();
    logger.info("Browser closed");
  }
}

process.on("exit", closeBrowser);
process.on("SIGINT", closeBrowser);

(async () => {
  await initializeBrowser();
  const url = process.argv[2];
  const selector = process.argv[3];

  if (!url || !selector) {
    logger.error("URL and selector must be provided as arguments.");
    process.exit(1);
  }

  const html = await scrapePage(url, selector);
  if (html) {
    logger.info("HTML content fetched successfully");
    console.log(html);
  } else {
    logger.error("Failed to fetch page content.");
  }

  await closeBrowser();
})();
