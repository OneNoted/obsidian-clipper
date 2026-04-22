import { applyGoFilterOrFallback } from './go-filter';

const fallbackCamel = (str: string): string => str
	.replace(/(?:^\w|[A-Z]|\b\w)/g, (letter, index) => 
		index === 0 ? letter.toLowerCase() : letter.toUpperCase()
	)
	.replace(/[\s_-]+/g, '');

export const camel = (str: string): string => {
	return applyGoFilterOrFallback('camel', str, undefined, () => fallbackCamel(str));
};
