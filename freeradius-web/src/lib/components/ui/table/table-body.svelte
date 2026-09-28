<script lang="ts">
	import { cn } from '$lib/utils.js';
	let klass: string | undefined = undefined;
	export { klass as class };
</script>

<tbody class={cn('[&_tr:last-child]:border-0', klass)} {...$$restProps}>
	<slot />
</tbody>
