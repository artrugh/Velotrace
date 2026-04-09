import { execSync, spawn } from "child_process";
import fs from "fs";
import path from "path";
import http from "http";

// --- Configuration ---
const APP_NAME = "web-portal";
const ROOT_DIR = process.cwd();
const APP_PATH = path.join(ROOT_DIR, "apps", "web-portal");
const ENV_PATH = path.join(APP_PATH, ".env");
const PORT = 3000;

// 0. Argument Check
const googleClientId = process.argv[2];
if (!googleClientId) {
  console.error("\x1b[31m❌ Error: GOOGLE_CLIENT_ID required.\x1b[0m");
  process.exit(1);
}

const isPortBusy = (port) => {
  return new Promise((resolve) => {
    const server = http
      .createServer()
      .once("error", () => resolve(true))
      .once("listening", () => {
        server.close();
        resolve(false);
      })
      .listen(port);
  });
};

const waitForPort = async (port, retries = 5, delay = 1000) => {
  for (let i = 0; i < retries; i++) {
    if (!(await isPortBusy(port))) return true;
    if (i < retries - 1) {
      console.log(
        `⏳ Port ${port} is still releasing, retrying... (${i + 1}/${retries})`,
      );
      await new Promise((r) => setTimeout(r, delay));
    }
  }
  return false;
};

async function run() {
  try {
    console.log(`🚀 Starting setup for ${APP_NAME}...`);

    // 1. Docker Cleanup
    try {
      const containerId = execSync(`docker ps -q --filter "publish=${PORT}"`)
        .toString()
        .trim();
      if (containerId) {
        console.log(`🛑 Stopping Docker container on port ${PORT}...`);
        execSync(`docker stop ${containerId}`);
      }
    } catch (e) {}

    // 2. Wait for Port
    if (!(await waitForPort(PORT))) {
      console.error(`\x1b[31m❌ Error: Port ${PORT} is busy.\x1b[0m`);
      process.exit(1);
    }

    // 3. Environment & File Sync
    console.log("📝 Syncing configuration and contracts...");
    fs.writeFileSync(
      ENV_PATH,
      `NODE_ENV=development\nNITRO_OUTPUT_DIR=.vercel/output\nGOOGLE_CLIENT_ID=${googleClientId}\nIDENTITY_API_URL=http://localhost:8080\nBIKES_API_URL=http://localhost:8081\n`,
    );

    const srcLib = path.join(ROOT_DIR, "libs", "api-contract");
    const destLib = path.join(APP_PATH, "libs", "api-contract");

    if (fs.existsSync(srcLib)) {
      if (fs.existsSync(destLib))
        fs.rmSync(destLib, { recursive: true, force: true });
      fs.cpSync(srcLib, destLib, { recursive: true, force: true });

      const appTools = path.join(APP_PATH, "tools");
      if (!fs.existsSync(appTools)) fs.mkdirSync(appTools, { recursive: true });
      fs.copyFileSync(
        path.join(ROOT_DIR, "tools", "generate-api-contracts.js"),
        path.join(appTools, "generate-api-contracts.js"),
      );
    }

    // 4. Install Dependencies
    // Use --recursive and --filter to ensure binaries are linked correctly in the monorepo
    console.log("📦 Installing dependencies...");
    execSync(`pnpm install --filter ${APP_NAME}...`, {
      stdio: "inherit",
      cwd: ROOT_DIR,
    });

    // 5. Launch
    console.log(`✨ Starting ${APP_NAME} via NX...`);
    // Added 'pnpm exec' before nx to ensure the environment path is set correctly
    const serve = spawn("pnpm", ["exec", "nx", "serve", APP_NAME], {
      stdio: "inherit",
      shell: true,
      cwd: ROOT_DIR,
    });

    serve.on("close", (code) => process.exit(code));
  } catch (error) {
    console.error(`\x1b[31m❌ Setup failed: ${error.message}\x1b[0m`);
    process.exit(1);
  }
}

run();
