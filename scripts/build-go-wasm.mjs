import { copyFileSync, existsSync, mkdirSync, rmSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { spawnSync } from 'node:child_process';

const root = resolve(import.meta.dirname, '..');
const outDir = join(root, 'src', 'go-wasm');
const wasmOut = join(outDir, 'clipper.wasm');

rmSync(outDir, { recursive: true, force: true });

if (process.env.SKIP_GO_WASM === '1' || process.env.SKIP_GO_WASM === 'true') {
	console.log('[go-wasm] skipped by SKIP_GO_WASM');
	process.exit(0);
}

mkdirSync(outDir, { recursive: true });

const goVersion = spawnSync('go', ['version'], { encoding: 'utf8' });
if (goVersion.status !== 0) {
	console.error('[go-wasm] Go toolchain not found; refusing to package stale generated assets.');
	process.exit(goVersion.status ?? 1);
}

const env = { ...process.env, GOOS: 'js', GOARCH: 'wasm' };
const build = spawnSync('go', ['build', '-ldflags=-s -w', '-o', wasmOut, './go/cmd/clipper-wasm'], {
	cwd: root,
	env,
	stdio: 'inherit'
});

if (build.status !== 0) {
	process.exit(build.status ?? 1);
}

const goroot = spawnSync('go', ['env', 'GOROOT'], { encoding: 'utf8' });
if (goroot.status !== 0) {
	console.error('[go-wasm] built clipper.wasm but could not locate GOROOT for wasm_exec.js; refusing to package incomplete assets.');
	process.exit(goroot.status ?? 1);
}

const goRootPath = goroot.stdout.trim();
const supportCandidates = [
	join(goRootPath, 'lib', 'wasm', 'wasm_exec.js'),
	join(goRootPath, 'misc', 'wasm', 'wasm_exec.js'),
];
const wasmExec = supportCandidates.find(existsSync);
if (!wasmExec) {
	console.error('[go-wasm] built clipper.wasm but wasm_exec.js was not found in the Go toolchain; refusing to package incomplete assets.');
	process.exit(1);
}

copyFileSync(wasmExec, join(outDir, 'wasm_exec.js'));
console.log(`[go-wasm] built ${wasmOut}`);
