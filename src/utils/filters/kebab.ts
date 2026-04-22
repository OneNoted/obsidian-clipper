import { applyGoFilterOrFallback } from './go-filter';

const fallbackKebab = (str: string): string => str
	.replace(/([a-z])([A-Z])/g, '$1-$2')
	.replace(/[\s_]+/g, '-')
	.toLowerCase();

export const kebab = (str: string): string => {
	return applyGoFilterOrFallback('kebab', str, undefined, () => fallbackKebab(str));
};
