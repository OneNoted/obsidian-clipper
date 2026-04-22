import { applyGoFilterOrFallback } from './go-filter';

export const decode_uri = (str: string): string => {
	return applyGoFilterOrFallback('decode_uri', str, undefined, () => {
		try {
			return decodeURIComponent(str);
		} catch {
			// If decoding fails (e.g., malformed URI), return the original string
			return str;
		}
	});
};
