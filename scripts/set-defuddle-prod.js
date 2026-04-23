const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

// Run commands from the project root
const projectRoot = path.join(__dirname, '..');
const packageJsonPath = path.join(projectRoot, 'package.json');
const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));

packageJson.dependencies.defuddle = '0.18.1';
fs.writeFileSync(packageJsonPath, JSON.stringify(packageJson, null, '\t') + '\n');

try {
	// Remove defuddle module and reinstall from the committed Bun lockfile.
	execSync('rm -rf node_modules/defuddle && bun install', { 
		stdio: 'inherit',
		cwd: projectRoot
	});
} catch (error) {
	console.error('Failed to update defuddle:', error);
	process.exit(1);
} 