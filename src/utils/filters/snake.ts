import { applyGoFilterOrFallback } from './go-filter';

const fallbackSnake = (str: string): string => str
	.replace(/([a-z])([A-Z])/g, '$1_$2')
	.replace(/[\s-]+/g, '_')
	.toLowerCase();

export const snake = (str: string): string => {
	return applyGoFilterOrFallback('snake', str, undefined, () => fallbackSnake(str));
};
