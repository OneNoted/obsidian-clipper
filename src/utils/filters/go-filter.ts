import { applyGoFilterSync, warmGoWasmRuntime } from '../go-wasm';

void warmGoWasmRuntime();

type Fallback<T> = () => T;

interface GoFilterOptions {
	avoidJsonObjects?: boolean;
	avoidJsonArraysWithObjects?: boolean;
	avoidJsonScalars?: boolean;
	avoidJsonValues?: boolean;
}

export function applyGoFilterOrFallback<T extends string>(
	filterName: string,
	input: string,
	param: string | undefined,
	fallback: Fallback<T>,
	options: GoFilterOptions = {}
): T {
	if (options.avoidJsonValues && looksLikeJson(input)) {
		return fallback();
	}
	if (options.avoidJsonObjects && looksLikeJsonObject(input)) {
		return fallback();
	}
	if (options.avoidJsonArraysWithObjects && looksLikeJsonArrayWithObjects(input)) {
		return fallback();
	}
	if (options.avoidJsonScalars && looksLikeJsonScalar(input)) {
		return fallback();
	}

	const goResult = applyGoFilterSync(filterName, input, param);
	return goResult === undefined ? fallback() : goResult as T;
}

export function looksLikeJson(input: string): boolean {
	const trimmed = input.trim();
	if (!trimmed) {
		return false;
	}

	try {
		JSON.parse(trimmed);
		return true;
	} catch {
		return false;
	}
}

export function looksLikeJsonObject(input: string): boolean {
	const trimmed = input.trim();
	return trimmed.startsWith('{') && trimmed.endsWith('}');
}

export function looksLikeJsonArrayWithObjects(input: string): boolean {
	const trimmed = input.trim();
	if (!trimmed.startsWith('[') || !trimmed.endsWith(']')) {
		return false;
	}

	try {
		const parsed = JSON.parse(trimmed);
		return Array.isArray(parsed) && parsed.some(item => typeof item === 'object' && item !== null);
	} catch {
		return false;
	}
}

export function looksLikeJsonScalar(input: string): boolean {
	const trimmed = input.trim();
	if (!trimmed || trimmed.startsWith('{') || trimmed.startsWith('[')) {
		return false;
	}

	try {
		JSON.parse(trimmed);
		return true;
	} catch {
		return false;
	}
}
