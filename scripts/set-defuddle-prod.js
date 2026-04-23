const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

// Run commands from the project root
const projectRoot = path.join(__dirname, '..');
const packageJsonPath = path.join(projectRoot, 'package.json');
const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
const prodDependency = packageJson.config?.defuddleProdDependency ?? packageJson.dependencies.defuddle;

if (!prodDependency || prodDependency.startsWith('file:')) {
	console.error('Unable to determine production defuddle dependency. Run from a clean checkout or set config.defuddleProdDependency.');
	process.exit(1);
}

packageJson.dependencies.defuddle = prodDependency;
if (packageJson.config) {
	delete packageJson.config.defuddleProdDependency;
	if (Object.keys(packageJson.config).length === 0) {
		delete packageJson.config;
	}
}
fs.writeFileSync(packageJsonPath, JSON.stringify(packageJson, null, '\t') + '\n');

try {
	// Remove defuddle module and reinstall from the Bun lockfile.
	execSync('rm -rf node_modules/defuddle && bun install', {
		stdio: 'inherit',
		cwd: projectRoot
	});
} catch (error) {
	console.error('Failed to update defuddle:', error);
	process.exit(1);
}
