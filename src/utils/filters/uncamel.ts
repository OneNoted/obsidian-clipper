import { applyGoFilterOrFallback } from './go-filter';

const fallbackUncamel = (str: string): string => str
	.replace(/([a-z0-9])([A-Z])/g, '$1 $2')
	.replace(/([A-Z])([A-Z][a-z])/g, '$1 $2')
	.toLowerCase();

export const uncamel = (str: string): string => {
	return applyGoFilterOrFallback('uncamel', str, undefined, () => fallbackUncamel(str));
};
