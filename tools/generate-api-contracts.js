import { execSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";

// --- Configuration ---
// Using path.resolve ensures we aren't tripped up by the Docker WORKDIR
const CONTRACTS_DIR = path.resolve("./libs/api-contract");
const OUTPUT_DIR = path.join(CONTRACTS_DIR, ".generated");
const CONVERT_CMD = "pnpm dlx swagger2openapi@7.0.8";
const GEN_CMD = "pnpm dlx openapi-typescript@7.13.0";

if (!fs.existsSync(OUTPUT_DIR)) {
  fs.mkdirSync(OUTPUT_DIR, { recursive: true });
}

const syncService = (swaggerFile) => {
  const serviceDir = path.dirname(swaggerFile);
  const serviceName = path.basename(serviceDir).replace("-api", "");

  const tempOpenApi = path.join(OUTPUT_DIR, `${serviceName}-openapi.yaml`);
  const outputFile = path.join(OUTPUT_DIR, `${serviceName}.ts`);

  try {
    console.log(`\x1b[36m🔄 Processing: ${serviceName}\x1b[0m`);

    // 1. Convert Swagger 2.0 to OpenAPI 3.0
    // Using 'inherit' so we see exactly what swagger2openapi is doing
    execSync(
      `${CONVERT_CMD} "${swaggerFile}" --outfile "${tempOpenApi}" --patch`,
      { stdio: "inherit" },
    );

    // 2. Generate TypeScript types
    execSync(`${GEN_CMD} "${tempOpenApi}" -o "${outputFile}"`, {
      stdio: "inherit",
    });

    // 3. Force LF line endings
    const content = fs.readFileSync(outputFile, "utf8");
    fs.writeFileSync(outputFile, content.replace(/\r\n/g, "\n"), "utf8");

    // --- REMOVED fs.unlinkSync here to prevent race conditions ---

    console.log(`\x1b[32m✅ Generated: ${outputFile}\x1b[0m`);
  } catch (err) {
    console.error(
      `\x1b[31m❌ Error syncing ${serviceName}: ${err.message}\x1b[0m`,
    );
  }
};

const findSwaggerFiles = (dir, fileList = []) => {
  const files = fs.readdirSync(dir);
  for (const file of files) {
    const filePath = path.join(dir, file);
    if (file === ".generated") continue;

    if (fs.statSync(filePath).isDirectory()) {
      findSwaggerFiles(filePath, fileList);
    } else if (file === "swagger.yaml") {
      fileList.push(filePath);
    }
  }
  return fileList;
};

const cleanupTempFiles = () => {
  const files = fs.readdirSync(OUTPUT_DIR);
  files.forEach((file) => {
    if (file.endsWith("-openapi.yaml")) {
      fs.unlinkSync(path.join(OUTPUT_DIR, file));
    }
  });
};

const syncAll = () => {
  console.log("\n--- 🚀 Starting API Contract Sync ---");
  const files = findSwaggerFiles(CONTRACTS_DIR);
  files.forEach(syncService);

  // Cleanup only happens once everything is done
  cleanupTempFiles();

  console.log("--- ✨ Sync Complete ---\n");
};

// --- Execution ---
const isWatchMode = process.argv.includes("--watch");

if (isWatchMode) {
  syncAll();
  console.log(`Watching for changes in ${CONTRACTS_DIR}...`);

  let timeout;
  fs.watch(CONTRACTS_DIR, { recursive: true }, (eventType, filename) => {
    if (
      filename?.endsWith("swagger.yaml") &&
      !filename.includes(".generated")
    ) {
      clearTimeout(timeout);
      timeout = setTimeout(() => {
        console.log(`\nDetected change in ${filename}...`);
        syncAll();
      }, 500);
    }
  });
} else {
  syncAll();
}
