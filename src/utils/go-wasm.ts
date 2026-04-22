type GoRuntimeConstructor = new () => {
	importObject: WebAssembly.Imports;
	run(instance: WebAssembly.Instance): Promise<void>;
};

type ObsidianClipperGoRuntime = {
	applyFilter(filterName: string, input: string, param?: string): string | null | undefined;
};

type GoGlobal = typeof globalThis & {
	Go?: GoRuntimeConstructor;
	obsidianClipperGo?: ObsidianClipperGoRuntime;
	__obsidianClipperGoRuntime?: Promise<boolean>;
};

const globalScope = globalThis as GoGlobal;

export function applyGoFilterSync(filterName: string, input: string, param?: string): string | undefined {
	const runtime = globalScope.obsidianClipperGo;
	if (!runtime?.applyFilter) {
		return undefined;
	}

	try {
		const result = runtime.applyFilter(filterName, input, param);
		return typeof result === 'string' ? result : undefined;
	} catch (error) {
		console.warn(`[go-wasm] ${filterName} failed; falling back to TypeScript`, error);
		return undefined;
	}
}

export function warmGoWasmRuntime(): Promise<boolean> {
	if (globalScope.__obsidianClipperGoRuntime) {
		return globalScope.__obsidianClipperGoRuntime;
	}

	globalScope.__obsidianClipperGoRuntime = loadGoWasmRuntime();
	return globalScope.__obsidianClipperGoRuntime;
}

async function loadGoWasmRuntime(): Promise<boolean> {
	if (globalScope.obsidianClipperGo?.applyFilter) {
		return true;
	}

	if (typeof WebAssembly === 'undefined' || typeof fetch === 'undefined') {
		return false;
	}

	// Keep content scripts and service workers on the TypeScript path. Loading
	// wasm_exec.js is straightforward and CSP-safe on extension pages, but page
	// content scripts execute in isolated worlds where script-tag bootstrapping is
	// browser-specific and should not block clipping.
	const documentRef = globalScope.document;
	if (!documentRef?.head || !isExtensionPage()) {
		return false;
	}

	try {
		await loadGoSupportScript(getExtensionUrl('go-wasm/wasm_exec.js'));
		if (!globalScope.Go) {
			return false;
		}

		const go = new globalScope.Go();
		const response = await fetch(getExtensionUrl('go-wasm/clipper.wasm'));
		if (!response.ok) {
			return false;
		}
		const bytes = await response.arrayBuffer();
		const result = await WebAssembly.instantiate(bytes, go.importObject);
		void go.run(result.instance).catch(error => {
			console.warn('[go-wasm] Go runtime stopped', error);
			delete globalScope.obsidianClipperGo;
			delete globalScope.__obsidianClipperGoRuntime;
		});

		return waitForRuntime();
	} catch (error) {
		console.warn('[go-wasm] unavailable; using TypeScript fallbacks', error);
		return false;
	}
}

function getExtensionUrl(path: string): string {
	const runtime = (globalScope as any).browser?.runtime ?? (globalScope as any).chrome?.runtime;
	return typeof runtime?.getURL === 'function' ? runtime.getURL(path) : path;
}

function isExtensionPage(): boolean {
	const protocol = globalScope.location?.protocol;
	return protocol === 'chrome-extension:' || protocol === 'moz-extension:' || protocol === 'safari-web-extension:';
}

function loadGoSupportScript(src: string): Promise<void> {
	if (globalScope.Go) {
		return Promise.resolve();
	}

	return new Promise((resolve, reject) => {
		const script = globalScope.document!.createElement('script');
		script.src = src;
		script.async = true;
		script.onload = () => resolve();
		script.onerror = () => reject(new Error(`Failed to load ${src}`));
		globalScope.document!.head.appendChild(script);
	});
}

function waitForRuntime(): Promise<boolean> {
	return new Promise(resolve => {
		let attempts = 0;
		const check = () => {
			if (globalScope.obsidianClipperGo?.applyFilter) {
				resolve(true);
				return;
			}
			attempts += 1;
			if (attempts > 50) {
				resolve(false);
				return;
			}
			setTimeout(check, 10);
		};
		check();
	});
}
