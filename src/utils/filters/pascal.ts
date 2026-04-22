import { applyGoFilterOrFallback } from './go-filter';

const fallbackPascal = (str: string): string => str
	.replace(/[\s_-]+(.)/g, (_, c) => c.toUpperCase())
	.replace(/^(.)/, c => c.toUpperCase());

export const pascal = (str: string): string => {
	return applyGoFilterOrFallback('pascal', str, undefined, () => fallbackPascal(str));
};
