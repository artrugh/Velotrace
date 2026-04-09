import { execSync, spawn } from "child_process";
import fs from "fs";
import path from "path";
import http from "http";

// --- Configuration ---
const APP_NAME = "web-portal";
const ROOT_DIR = process.cwd();
const APP_PATH = path.join(ROOT_DIR, "apps", "web-portal");
const ENV_PATH = path.join(APP_PATH, ".env");
const ROOT_ENV_PATH = path.join(ROOT_DIR, ".env");

/**
 * Helper to parse .env files manually
 */
const parseEnv = (filePath) => {
  if (!fs.existsSync(filePath)) {
    console.error(`\x1b[31m❌ Error: Root .env file not found at ${filePath}\x1b[0m`);
    process.exit(1);
  }
  const content = fs.readFileSync(filePath, "utf8");
  const env = {};
  content.split("\n").forEach((line) => {
    const [key, value] = line.split("=");
    if (key && value) env[key.trim()] = value.trim();
  });
  return env;
};

// 0. Load Configuration
const rootEnv = parseEnv(ROOT_ENV_PATH);
const WEB_PORTAL_PORT = rootEnv.WEB_PORTAL_PORT || 3000;
const IDENTITY_API_PORT = rootEnv.IDENTITY_API_PORT || 8080;
const BIKES_API_PORT = rootEnv.BIKES_API_PORT || 8081;

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
      console.log(`⏳ Port ${port} is still releasing, retrying... (${i + 1}/${retries})`);
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
      const containerId = execSync(`docker ps -q --filter "publish=${WEB_PORTAL_PORT}"`)
        .toString()
        .trim();
      if (containerId) {
        console.log(`🛑 Stopping Docker container on port ${WEB_PORTAL_PORT}...`);
        execSync(`docker stop ${containerId}`);
      }
    } catch (e) {}

    // 2. Wait for Port
    if (!(await waitForPort(WEB_PORTAL_PORT))) {
      console.error(`\x1b[31m❌ Error: Port ${WEB_PORTAL_PORT} is busy.\x1b[0m`);
      process.exit(1);
    }

    // 3. Environment & Symlink Sync
    console.log("📝 Wiring configuration and contracts...");
    const portalEnvContent = [
      `NODE_ENV=development`,
      `NITRO_OUTPUT_DIR=.vercel/output`,
      `GOOGLE_CLIENT_ID=${googleClientId}`,
      `IDENTITY_API_URL=http://localhost:${IDENTITY_API_PORT}`,
      `BIKES_API_URL=http://localhost:${BIKES_API_PORT}`,
      `NUXT_PORT=${WEB_PORTAL_PORT}`
    ].join("\n");

    fs.writeFileSync(ENV_PATH, portalEnvContent);

  const srcLib = path.join(ROOT_DIR, "libs", "api-contract");
    const destLib = path.join(APP_PATH, "libs", "api-contract");

    if (fs.existsSync(srcLib)) {
      // 1. Handle the Library Folder (Junction works without Admin)
      const destLibParent = path.dirname(destLib);
      if (!fs.existsSync(destLibParent)) fs.mkdirSync(destLibParent, { recursive: true });

      if (fs.existsSync(destLib)) {
        // Check if it's already a link or a directory and remove it
        fs.rmSync(destLib, { recursive: true, force: true });
      }
      
      fs.symlinkSync(srcLib, destLib, "junction");
      console.log("🔗 Symlinked libs/api-contract (Junction).");

      // 2. Handle the Generator Tool (Copy instead of Symlink to avoid EPERM)
      const appTools = path.join(APP_PATH, "tools");
      if (!fs.existsSync(appTools)) fs.mkdirSync(appTools, { recursive: true });

      const srcTool = path.join(ROOT_DIR, "tools", "generate-api-contracts.js");
      const destTool = path.join(appTools, "generate-api-contracts.js");
      
      if (fs.existsSync(destTool)) fs.rmSync(destTool, { force: true });
      
      // Use copyFileSync here because Windows treats file symlinks differently than directory junctions
      fs.copyFileSync(srcTool, destTool);
      console.log("✅ Copied generate-api-contracts.js to local tools.");
    }

    // 4. Install Dependencies
    console.log("📦 Installing dependencies...");
    execSync(`pnpm install --filter ${APP_NAME}...`, { stdio: "inherit", cwd: ROOT_DIR });

    // 5. Launch via NX with Robust Error Handling
    console.log(`✨ Starting ${APP_NAME} via NX...`);
    const serve = spawn("pnpm", ["exec", "nx", "serve", APP_NAME], {
      stdio: "inherit",
      shell: true,
      cwd: ROOT_DIR,
    });

    serve.on("error", (error) => {
      console.error(`\x1b[31m❌ Failed to launch ${APP_NAME}: ${error.message}\x1b[0m`);
      process.exit(1);
    });

    serve.on("close", (code, signal) => {
      // Exit with 1 if it crashed or was killed by a signal, otherwise 0
      process.exit(code ?? (signal ? 1 : 0));
    });

  } catch (error) {
    console.error(`\x1b[31m❌ Setup failed: ${error.message}\x1b[0m`);
    process.exit(1);
  }
}

run();