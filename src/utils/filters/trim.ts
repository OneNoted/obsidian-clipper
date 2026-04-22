import { applyGoFilterOrFallback } from './go-filter';

export const trim = (str: string): string => {
	return applyGoFilterOrFallback('trim', str, undefined, () => str.trim());
};
