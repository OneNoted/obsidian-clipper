import { applyGoFilterOrFallback } from './go-filter';

export const upper = (input: string | string[]): string | string[] => {
	const toUpperCase = (str: string): string => {
		return str.toLocaleUpperCase();
	};

	if (Array.isArray(input)) {
		return input.map(toUpperCase);
	}

	return applyGoFilterOrFallback('upper', input, undefined, () => toUpperCase(input));
};
