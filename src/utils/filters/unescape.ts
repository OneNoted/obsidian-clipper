import { applyGoFilterOrFallback } from './go-filter';

const fallbackUnescape = (str: string): string => str
	.replace(/\\"/g, '"')
	.replace(/\\n/g, '\n');

export const unescape = (str: string): string => {
	return applyGoFilterOrFallback('unescape', str, undefined, () => fallbackUnescape(str));
};
