const express = require("express");
const puppeteer = require("puppeteer");
const winston = require("winston");

const app = express();
app.use(express.json());

const PORT = 3000;
const POOL_SIZE = 5;

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

class BrowserPool {
  constructor(poolSize) {
    this.pool = [];
    this.poolSize = poolSize;
  }

  async initialize() {
    for (let i = 0; i < this.poolSize; i++) {
      const browser = await puppeteer.launch();
      this.pool.push(browser);
    }
    logger.info(`Browser pool initialized with ${this.poolSize} instances`);
  }

  async getBrowser() {
    if (this.pool.length > 0) {
      return this.pool.pop();
    }
    logger.warn("All browsers in use, creating new instance");
    return await puppeteer.launch();
  }

  releaseBrowser(browser) {
    if (this.pool.length < this.poolSize) {
      this.pool.push(browser);
    } else {
      browser.close();
    }
  }

  async closeAll() {
    await Promise.all(this.pool.map((browser) => browser.close()));
    this.pool = [];
    logger.info("All browser instances closed");
  }
}

const browserPool = new BrowserPool(POOL_SIZE);

app.post("/scrape", async (req, res) => {
  const { url, selector } = req.body;
  if (!url || !selector) {
    return res.status(400).json({ error: "URL and selector must be provided" });
  }

  const browser = await browserPool.getBrowser();
  try {
    const page = await browser.newPage();
    logger.info(`Navigating to ${url}`);
    await page.goto(url, { waitUntil: "networkidle2", timeout: 60000 });
    logger.info(`Waiting for selector ${selector}`);
    await page.waitForSelector(selector, { timeout: 60000 });
    const html = await page.content();
    logger.info("Page content fetched");
    await page.close();
    res.json({ html });
  } catch (error) {
    logger.error(`Error: ${error.message}`);
    res.status(500).json({ error: error.message });
  } finally {
    browserPool.releaseBrowser(browser);
  }
});

async function startServer() {
  await browserPool.initialize();
  app.listen(PORT, () => {
    logger.info(`Puppeteer service listening on port ${PORT}`);
  });
}

startServer();

process.on("SIGINT", async () => {
  logger.info("Shutting down...");
  await browserPool.closeAll();
  process.exit();
});
