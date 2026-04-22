import { applyGoFilterOrFallback } from './go-filter';

export const lower = (input: string | string[]): string | string[] => {
	const toLowerCase = (str: string): string => {
		return str.toLocaleLowerCase();
	};

	if (Array.isArray(input)) {
		return input.map(toLowerCase);
	}

	return applyGoFilterOrFallback('lower', input, undefined, () => toLowerCase(input));
};
