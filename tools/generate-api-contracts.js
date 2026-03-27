import { execSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';


// --- Configuration ---
const CONTRACTS_DIR = './libs/api-contract';
const OUTPUT_DIR = path.join(CONTRACTS_DIR, '.generated'); // Centralized folder
const CONVERT_CMD = 'pnpm dlx swagger2openapi@7.0.8';
const GEN_CMD = 'pnpm dlx openapi-typescript@7.13.0';

/**
 * Ensures the output directory exists
 */
if (!fs.existsSync(OUTPUT_DIR)) {
  fs.mkdirSync(OUTPUT_DIR, { recursive: true });
}

/**
 * Syncs a single service into the central directory
 */
const syncService = (swaggerFile) => {
  const serviceDir = path.dirname(swaggerFile);
  const serviceName = path.basename(serviceDir).replace('-api', '');
  
  // We'll put the intermediate openapi.yaml in the central dir too
  const tempOpenApi = path.join(OUTPUT_DIR, `${serviceName}-openapi.yaml`);
  const outputFile = path.join(OUTPUT_DIR, `${serviceName}.ts`);

  try {
    console.log(`\x1b[36m🔄 Processing: ${serviceName}\x1b[0m`);

    // 1. Convert Swagger 2.0 to OpenAPI 3.0
    execSync(`${CONVERT_CMD} "${swaggerFile}" --outfile "${tempOpenApi}" --patch`, { stdio: 'pipe' });

    // 2. Generate TypeScript types
    execSync(`${GEN_CMD} "${tempOpenApi}" -o "${outputFile}"`, { stdio: 'pipe' });

    // 3. Force LF line endings
    const content = fs.readFileSync(outputFile, 'utf8');
    fs.writeFileSync(outputFile, content.replace(/\r\n/g, '\n'), 'utf8');

    // 4. Cleanup temporary YAML
    if (fs.existsSync(tempOpenApi)) fs.unlinkSync(tempOpenApi);

    console.log(`\x1b[32m✅ Generated: ${outputFile}\x1b[0m`);
  } catch (err) {
    console.error(`\x1b[31m❌ Error syncing ${serviceName}: ${err.message}\x1b[0m`);
  }
};

/**
 * Find all swagger.yaml files
 */
const findSwaggerFiles = (dir, fileList = []) => {
  const files = fs.readdirSync(dir);
  for (const file of files) {
    const filePath = path.join(dir, file);
    if (file === '.generated') continue; // Skip the output folder itself
    
    if (fs.statSync(filePath).isDirectory()) {
      findSwaggerFiles(filePath, fileList);
    } else if (file === 'swagger.yaml') {
      fileList.push(filePath);
    }
  }
  return fileList;
};

const syncAll = () => {
  console.log('\n--- 🚀 Starting API Contract Sync ---');
  const files = findSwaggerFiles(CONTRACTS_DIR);
  files.forEach(syncService);
  console.log('--- ✨ Sync Complete ---\n');
};

// --- Execution ---
const isWatchMode = process.argv.includes('--watch');

if (isWatchMode) {
  syncAll();
  console.log(`Watching for changes in ${CONTRACTS_DIR}...`);
  
  let timeout;
  fs.watch(CONTRACTS_DIR, { recursive: true }, (eventType, filename) => {
    // Only trigger if a swagger.yaml is changed and it's not in the output dir
    if (filename?.endsWith('swagger.yaml') && !filename.includes('.generated')) {
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