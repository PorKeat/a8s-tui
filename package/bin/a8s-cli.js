#!/usr/bin/env node

const { spawnSync } = require('child_process');
const path = require('path');

const archMap = { x64: 'amd64', arm64: 'arm64' };
const osMap = { darwin: 'darwin', linux: 'linux', win32: 'windows' };

const os = osMap[process.platform];
const arch = archMap[process.arch];

if (!os || !arch) {
  console.error(`Unsupported platform: ${process.platform} ${process.arch}`);
  process.exit(1);
}

const ext = os === 'windows' ? '.exe' : '';
const binName = `a8s-cli_${os}_${arch}${ext}`;
const binPath = path.join(__dirname, binName);

const result = spawnSync(binPath, process.argv.slice(2), { stdio: 'inherit' });

if (result.error) {
  console.error(`Failed to execute binary: ${binPath}`);
  console.error(result.error.message);
  process.exit(1);
}

process.exit(result.status);
