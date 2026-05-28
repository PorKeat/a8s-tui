const fs = require('fs');
const path = require('path');
const axios = require('axios');
const tar = require('tar');
const unzipper = require('unzipper');

const osMap = {
  darwin: 'darwin',
  linux: 'linux',
  win32: 'windows',
};
const archMap = {
  x64: 'amd64',
  arm64: 'arm64',
};

const repo = 'ITProfessional-Gen01/a8s-cli';
const binDir = path.join(__dirname, '..', 'bin');
const binPath = path.join(binDir, process.platform === 'win32' ? 'a8s-cli.exe' : 'a8s-cli');

async function install() {
  const os = osMap[process.platform];
  const arch = archMap[process.arch];

  if (!os || !arch) {
    console.error(`Unsupported platform: ${process.platform} ${process.arch}`);
    process.exit(1);
  }

  const pkg = require('../package.json');
  let version = pkg.version;
  if (version === '0.0.0-development') {
    console.log('Development version detected. Skipping binary download.');
    return;
  }

  const isWin = os === 'windows';
  const ext = isWin ? 'zip' : 'tar.gz';
  const filename = `a8s-cli_${os}_${arch}.${ext}`;
  const url = `https://github.com/${repo}/releases/download/v${version}/${filename}`;

  console.log(`Downloading a8s-cli v${version} for ${os}-${arch}...`);
  
  try {
    fs.mkdirSync(binDir, { recursive: true });
    
    const response = await axios({
      method: 'GET',
      url: url,
      responseType: 'stream',
    });

    if (isWin) {
      response.data.pipe(unzipper.Extract({ path: binDir }));
    } else {
      response.data.pipe(tar.x({ C: binDir }));
    }

    response.data.on('end', () => {
      setTimeout(() => {
        if (!isWin && fs.existsSync(binPath)) {
          fs.chmodSync(binPath, 0o755);
        }
        console.log('Installation complete!');
      }, 500); // Give streams a moment to close
    });
  } catch (error) {
    console.error(`Failed to download binary from ${url}`);
    console.error(error.message);
    process.exit(1);
  }
}

install();
